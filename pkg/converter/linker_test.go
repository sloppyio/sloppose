package converter_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	sloppy "github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/sloppose/pkg/converter"
)

func TestLinker_FindService(t *testing.T) {
	cases := map[string]struct {
		value       string
		shouldMatch bool
	}{
		"FOO":      {"bar", false},
		"BAR":      {"foo:80", true},
		"SHORT":    {"s:80", true},
		"FOO_BAR":  {"another.foo:443", true},
		"FOO_HOST": {"bar", true},
		"FOO_URL":  {"mongodb://whatever:4444/db", true},
		"BAR_URL":  {"foo://nope", true},
	}

	l := converter.Linker{}

	for envKey, caseVal := range cases {
		match := l.FindServiceString(envKey, caseVal.value)
		if caseVal.shouldMatch && match == "" {
			t.Errorf("Expected an match for %q, got nothing.", caseVal.value)
		} else if match != "" && !caseVal.shouldMatch {
			t.Errorf("Expected no match for %q, got: %v", caseVal.value, match)
		}
	}
}

func TestLinker_Resolve(t *testing.T) {
	linker := &converter.Linker{}
	name := "sloppy-test"

	expected := &converter.SloppyFile{
		Version: "v1",
		Project: name,
		Services: map[string]converter.SloppyApps{
			"apps": {
				"a": &converter.SloppyApp{
					App: &sloppy.App{
						Dependencies: []string{"../apps/b"},
						EnvVars: map[string]string{
							"API_AUTH": "some-external.service:80",
							"API_URL":  fmt.Sprintf("b.apps.%s:8080", name),
						},
						Image: ToStrPtr("hugo"),
					},
					Env: converter.SloppyEnvSlice{
						{"API_AUTH": "some-external.service:80"},
						{"API_URL": fmt.Sprintf("b.apps.%s:8080", name)},
					},
				},
				"b": &converter.SloppyApp{
					App: &sloppy.App{
						Image: ToStrPtr("golang"),
					},
					Port: ToIntPtr(8080),
				},
			},
		},
	}

	cf, sf := loadSloppyFile("/testdata/fixture_linker0.yml")
	err := linker.Resolve(cf, sf)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(sf, expected); diff != "" {
		t.Errorf("Result differs: (-got +want)\n%s", diff)
	}

}

func ToIntPtr(i int) *int {
	return &i
}

func ToStrPtr(s string) *string {
	return &s
}
