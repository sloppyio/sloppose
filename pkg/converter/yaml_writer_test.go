package converter_test

import (
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sloppyio/sloppose/internal/test"
	"github.com/sloppyio/sloppose/pkg/converter"
)

var updateFlag = flag.Bool("update", false, "go test -update")

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

func TestYAMLWriter_WriteFile(t *testing.T) {
	helper := test.NewHelper(t)
	_, sf := loadSloppyFile("testdata/docker-compose-v2.yml")
	writer := &converter.YAMLWriter{}
	helper.ChdirTemp()
	err := writer.WriteFile(sf, "test-tmp")
	helper.Must(err)

	haveBuf, err := ioutil.ReadFile("test-tmp.yml")
	helper.Must(err)

	helper.ChdirTest()

	if *updateFlag {
		err := writer.WriteFile(sf, "testdata/golden0.yml")
		helper.Must(err)
		t.Log("Successfully written updated golden file.")
	}

	goldenBuf, err := ioutil.ReadFile("testdata/golden0.yml")
	helper.Must(err)

	if diff := cmp.Diff(strings.Split(string(haveBuf), "\n"), strings.Split(string(goldenBuf), "\n")); diff != "" {
		t.Errorf("Result differs:\n%s", diff)
	}
}
