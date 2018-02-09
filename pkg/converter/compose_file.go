package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/libcompose/config"
	"github.com/ghodss/yaml"
)

const (
	DefaultProjectName    = "sloppyio"
	EnvComposeProjectName = "COMPOSE_PROJECT_NAME"
)

type ComposeFile struct {
	ProjectName    string
	ServiceConfigs *config.ServiceConfigs
}

type composeVersion struct {
	Version string `yaml:"version"`
}

func NewComposeFile(buf [][]byte, projectName string) (cf *ComposeFile, err error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("At least one readed file is required")
	}

	composeVersion, err := cf.parseVersion(buf[0])
	if err != nil {
		return nil, err
	}
	if len(buf) > 1 {
		err = cf.validateVersions(buf)
		if err != nil {
			return nil, err
		}
	}

	loader := &ComposeLoader{}
	switch strings.Split(composeVersion, ".")[0] {
	case "3":
		cf, err = loader.LoadVersion3(buf)
	case "2":
		cf, err = loader.LoadVersion2(buf)
	default:
		err = fmt.Errorf("missing version declaration in compose file")
	}
	if err != nil {
		return
	}

	if projectName != "" {
		cf.ProjectName = projectName
	} else {
		if cf.ProjectName == "" {
			if env, ok := os.LookupEnv(EnvComposeProjectName); ok {
				cf.ProjectName = env
			} else {
				cf.ProjectName, err = cf.newProjectName()
				if err != nil {
					return
				}
			}
		}
	}
	return
}

// Returns the current working directory name.
func (cf *ComposeFile) newProjectName() (p string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return p, err
	}
	p, err = filepath.Abs(wd)
	p = filepath.Base(p)

	if p == "." {
		p = DefaultProjectName
	}
	return
}

func (cf *ComposeFile) parseVersion(bytes []byte) (string, error) {
	var version composeVersion
	err := yaml.Unmarshal(bytes, &version)
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

func (cf *ComposeFile) validateVersions(in [][]byte) error {
	version, err := cf.parseVersion(in[0])
	if err != nil {
		return err
	}
	for _, bytes := range in[1:] {
		parsed, err := cf.parseVersion(bytes)
		if err != nil {
			return err
		}
		if version != parsed {
			return fmt.Errorf("docker-compose version mismatch, want version: %s", version)
		}
	}
	return nil
}
