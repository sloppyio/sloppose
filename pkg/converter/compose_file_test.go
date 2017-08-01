package converter_test

import (
	"testing"

	"github.com/sloppyio/sloppose/pkg/converter"
)

func TestNewComposeV2File(t *testing.T) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-v2.yml")
	if err != nil {
		t.Error(err)
	}

	cf, err := converter.NewComposeFile([][]byte{b}, "")
	if err != nil {
		t.Error(err)
	}

	services := []string{"busy_env", "wordpress", "db"}
	for _, service := range services {
		_, found := cf.ServiceConfigs.Get(service)
		if !found {
			t.Errorf("Couldn't find service %q", service)
		}
	}
}

func TestNewComposeV3File(t *testing.T) {
	reader := &converter.ComposeReader{}
	b, err := reader.Read("/testdata/docker-compose-v3.yml")
	if err != nil {
		t.Error(err)
	}

	cf, err := converter.NewComposeFile([][]byte{b}, "")
	if err != nil {
		t.Error(err)
	}

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
		t.Error(err)
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

func TestNewComposeFiles(t *testing.T) {
	reader := &converter.ComposeReader{}
	buf, err := reader.ReadAll([]string{"/testdata/docker-compose-v2.yml", "/testdata/docker-compose-v3.yml"})
	if err != nil {
		t.Error(err)
	}
	_, err = converter.NewComposeFile(buf, "")
	if err == nil {
		t.Errorf("Expected an error due to compose version mismatch.")
	}
}
