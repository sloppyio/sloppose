package converter

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type YAMLWriter struct {}

func (w *YAMLWriter) WriteFile(i interface{}, path string) error {
	bytes, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}