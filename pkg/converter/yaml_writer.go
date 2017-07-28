package converter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

type YAMLWriter struct{}

func (w *YAMLWriter) WriteFile(i interface{}, path string) error {
	bytes, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	path = w.ensureFileEnding(path)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	path, err = filepath.Rel(cwd, filepath.Join(cwd, path))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}

func (w *YAMLWriter) ensureFileEnding(path string) string {
	if strings.HasSuffix(path, ".yml") {
		return path
	}
	return path + ".yml"
}
