package converter_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	sloppy "github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/sloppose/pkg/converter"
)

var cfA *converter.ComposeFile
var sfA *converter.SloppyFile

const testProjectName = "linker_test"

func init() {
	reader := &converter.ComposeReader{}
	buf, err := reader.Read("/testdata/fixture_linker_a.yml")
	if err != nil {
		panic(err)
	}
	loader := &converter.ComposeLoader{}
	cfA, err = loader.LoadVersion2([][]byte{buf})
	cfA.ProjectName = testProjectName
	if err != nil {
		panic(err)
	}
	sfA, err = converter.NewSloppyFile(cfA)
	if err != nil {
		panic(err)
	}
}

func TestLinker_Resolve(t *testing.T) {
	linker := &converter.Linker{}

	expected := &converter.SloppyFile{
		Version: "v1",
		Project: testProjectName,
		Services: map[string]converter.SloppyApps{
			"apps": {
				"a": &converter.SloppyApp{
					App: &sloppy.App{
						Dependencies: []string{"../apps/b"},
						Domain:       &sloppy.Domain{URI: ToStrPtr(converter.DomainUri)},
						EnvVars: map[string]string{
							"API_AUTH": "some-external.service:80",
							"API_URL":  fmt.Sprintf("b.apps.%s:8080", testProjectName),
						},
						Image:     ToStrPtr("hugo"),
						Instances: ToIntPtr(converter.InstanceCount),
						Memory:    ToIntPtr(converter.InstanceMemory),
					},
					Domain: converter.DomainUri,
					Env: []string{
						"API_AUTH=some-external.service:80",
						fmt.Sprintf("API_URL=b.apps.%s:8080", testProjectName),
					},
				},
				"b": &converter.SloppyApp{
					App: &sloppy.App{
						Domain:       &sloppy.Domain{URI: ToStrPtr(converter.DomainUri)},
						Image:        ToStrPtr("golang"),
						Instances:    ToIntPtr(converter.InstanceCount),
						Memory:       ToIntPtr(converter.InstanceMemory),
						PortMappings: []*sloppy.PortMap{{ToIntPtr(8080)}},
					},
					Domain: converter.DomainUri,
					Port:   ToIntPtr(8080),
				},
			},
		},
	}

	err := linker.Resolve(cfA, sfA)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(sfA, expected); diff != "" {
		t.Errorf("Result differs: (-got +want)\n%s", diff)
	}

}

func ToIntPtr(i int) *int {
	return &i
}

func ToStrPtr(s string) *string {
	return &s
}
