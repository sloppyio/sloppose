package converter_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sloppyio/sloppose/internal/test"
	"github.com/sloppyio/sloppose/pkg/converter"
)

func TestNewComposeV2File(t *testing.T) {
	helper := test.NewHelper(t)
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-v2.yml")
	helper.Must(err)

	cf, err := converter.NewComposeFile([][]byte{b}, "")
	helper.Must(err)

	services := []string{"busy_env", "wordpress", "db"}
	for _, service := range services {
		_, found := cf.ServiceConfigs.Get(service)
		if !found {
			t.Errorf("Couldn't find service %q", service)
		}
	}
}

func TestNewComposeV3File(t *testing.T) {
	helper := test.NewHelper(t)
	r := helper.GetTestFile("docker-compose-v3-full.yml")
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	helper.Must(err)

	cf, err := converter.NewComposeFile([][]byte{b}, "")
	helper.Must(err)

	services := []string{"foo"}
	for _, service := range services {
		_, found := cf.ServiceConfigs.Get(service)
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

	_, err = converter.NewComposeFile([][]byte{b}, "")
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
	b, err := reader.Read("/testdata/docker-compose-v2.yml")
	helper.Must(err)

	projectName := "myVeryCustomFooName"
	os.Setenv(converter.EnvComposeProjectName, projectName)
	defer os.Unsetenv(converter.EnvComposeProjectName)

	cf, err := converter.NewComposeFile([][]byte{b}, "")
	helper.Must(err)

	if cf.ProjectName != strings.ToLower(projectName) {
		t.Errorf("Expected %q as project name, got: %q", projectName, cf.ProjectName)
	}
}

func TestNewComposeFiles(t *testing.T) {
	reader := &converter.ComposeReader{}
	buf, err := reader.ReadAll([]string{"/testdata/docker-compose-v2.yml", "/testdata/docker-compose-v3.yml"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = converter.NewComposeFile(buf, "")
	if err == nil {
		t.Errorf("Expected an error due to compose version mismatch.")
	}
}
