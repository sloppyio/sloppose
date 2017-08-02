package converter_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/sloppyio/sloppose/pkg/converter"
)

// Tests the fallback to docker-compose.yml in current working dir.
// Since there is no one in current test wd, we expect an os error.
func TestComposeReader_ReadAllDefault(t *testing.T) {
	reader := &converter.ComposeReader{}
	_, err := reader.ReadAll([]string{})
	if (err != nil && reflect.TypeOf(&os.PathError{}) != reflect.TypeOf(err)) || err == nil {
		t.Errorf("Expected an os.PathError. Got: %t", err)
	}
}
