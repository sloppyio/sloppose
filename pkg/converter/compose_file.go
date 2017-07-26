package converter

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

const defaultFileName = "docker-compose.yml"

type ComposeFile struct {
	ServiceConfigs *config.ServiceConfigs
}

func NewComposeFile(files []string) (*ComposeFile, error) {
	cf, buf := &ComposeFile{}, [][]byte{}
	if len(files) > 0 {
		for _, file := range files {
			b, err := cf.readFromFile(file)
			if err != nil {
				return nil, err
			}
			buf = append(buf, b)
		}
	} else {
		b, err := cf.readFromFile(defaultFileName)
		if err != nil {
			return nil, err
		}
		buf = append(buf, b)
	}

	p := project.NewProject(&project.Context{
		ComposeBytes: buf,
	}, nil, nil)
	err := p.Parse()
	if err != nil {
		return nil, err
	}

	cfg := *p.ServiceConfigs
	cf.ServiceConfigs = &cfg

	return cf, nil
}

// Reads compose v2 bytes
// TODO: v3 support
// - implement hack for v3 (or bridge: https://github.com/aanand/compose-file)
// - https://github.com/docker/libcompose/issues/421
func (cf *ComposeFile) readFromFile(filename string) ([]byte, error) {
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filepath.Join(pwd, filename))
	if err != nil {
		return nil, err
	}

	return bytes, nil
}