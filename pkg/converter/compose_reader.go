package converter

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultFileName = "docker-compose.yml"
)

type ComposeReader struct{}

func (cr *ComposeReader) Read(filename string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path, err := filepath.Rel(cwd, filepath.Join(cwd, filename))
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (cr *ComposeReader) ReadAll(fileNames []string) (b [][]byte, err error) {
	files := fileNames[:]
	if len(files) == 0 {
		files = append(files, defaultFileName)
	}
	for _, file := range files {
		bytes, err := cr.Read(file)
		if err != nil {
			return nil, err
		}
		b = append(b, bytes)
	}
	return
}
