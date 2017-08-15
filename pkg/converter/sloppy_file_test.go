package converter_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sloppyio/sloppose/internal/test"
	"github.com/sloppyio/sloppose/pkg/converter"
)

// output should be the same as described above
var testFiles = []string{
	"docker-compose-v2.yml",
	"docker-compose-v3.yml",
}

func loadSloppyFile(filename string) (cf *converter.ComposeFile, sf *converter.SloppyFile) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read(filename)
	if err != nil {
		panic(err)
	}
	cf, err = converter.NewComposeFile([][]byte{b}, "sloppy-test")
	if err != nil {
		panic(err)
	}

	sf, err = converter.NewSloppyFile(cf)
	if err != nil {
		panic(err)
	}
	linker := &converter.Linker{}
	err = linker.Resolve(cf, sf)
	if err != nil {
		panic(err)
	}
	return
}

func TestNewSloppyFile(t *testing.T) {
	helper := test.NewHelper(t)
	expectedSloppyYml := helper.GetTestFile("golden0.yml")
	defer expectedSloppyYml.Close()
	b, err := ioutil.ReadAll(expectedSloppyYml)
	helper.Must(err)
	wantLines := strings.Split(string(b), "\n")

	for i, testFile := range testFiles {
		_, have := loadSloppyFile("testdata/" + testFile)

		helper.ChdirTemp()
		writer := &converter.YAMLWriter{}
		outFileName := fmt.Sprintf("out-%d.yml", i)
		err := writer.WriteFile(have, outFileName)
		helper.Must(err)

		haveBuf, err := ioutil.ReadFile(outFileName)
		helper.Must(err)

		haveLines := strings.Split(string(haveBuf), "\n")
		if diff := cmp.Diff(haveLines, wantLines); diff != "" {
			t.Errorf("Case: %q\nResult differs: (-got +want)\n%s", testFile, diff)
		}
		helper.ChdirTest()
	}
}

func TestNewSloppyFileInvalidPorts(t *testing.T) {
	helper := test.NewHelper(t)
	cases := []string{
		"fixture_sloppy-file0.yml",
		"fixture_sloppy-file1.yml",
	}

	for _, testCase := range cases {
		file := helper.GetTestFile(testCase)
		bytes, err := ioutil.ReadAll(file)
		file.Close()
		helper.Must(err)
		cf, err := converter.NewComposeFile([][]byte{bytes}, "")
		helper.Must(err)
		_, err = converter.NewSloppyFile(cf)
		if err == nil {
			t.Errorf("Expected a port related error, got nothing.")
		}
	}
}
