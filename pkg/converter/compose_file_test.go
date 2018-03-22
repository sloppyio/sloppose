package converter_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sloppyio/sloppose/internal/test"
	"github.com/sloppyio/sloppose/pkg/converter"
)

func TestNewComposeV3File(t *testing.T) {
	helper := test.NewHelper(t)
	r := helper.GetTestFile("docker-compose-v3-full.yml")
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	helper.Must(err)

	cf, err := converter.NewComposeFile(b, "")
	helper.Must(err)

	services := []string{"foo"}
	for _, service := range services {
		_, found := cf.ServiceConfigs[service]
		if !found {
			t.Errorf("Couldn't find service %q", service)
		}
	}
}

func TestNewComposeV3dot0File(t *testing.T) {
	helper := test.NewHelper(t)
	r := helper.GetTestFile("docker-compose-v3-0.yml")
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	helper.Must(err)

	cf, err := converter.NewComposeFile(b, "")
	helper.Must(err)

	services := []string{"busy_env"}
	for _, service := range services {
		_, found := cf.ServiceConfigs[service]
		if !found {
			t.Errorf("Couldn't find service %q", service)
		}
	}
}

func TestNewComposeVersionFile(t *testing.T) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-version.yml")
	if err != nil {
		t.Fatal(err)
	}

	_, err = converter.NewComposeFile(b, "")
	if err == nil {
		t.Errorf("Expected an error due to missing version field.")
	}
}

func TestNewComposeNilBytes(t *testing.T) {
	cf, err := converter.NewComposeFile(nil, "")
	if cf != nil && err == nil {
		t.Errorf("Expected an error due to zero bytes given.")
	}
}

func TestNewComposeFileProjectName(t *testing.T) {
	helper := test.NewHelper(t)
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-v3.yml")
	helper.Must(err)

	projectName := "myVeryCustomFooName"
	os.Setenv(converter.EnvComposeProjectName, projectName)
	defer os.Unsetenv(converter.EnvComposeProjectName)

	cf, err := converter.NewComposeFile(b, "")
	helper.Must(err)

	if diff := cmp.Diff(cf.ProjectName, strings.ToLower(projectName)); diff != "" {
		t.Errorf("Expected %q as project name, diff:\n%s", projectName, diff)
	}
}
