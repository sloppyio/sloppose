package converter

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/config"
	"github.com/ghodss/yaml"
)

const (
	defaultVersion        = "2"
	defaultProjectName    = "sloppyio"
	envComposeProjectName = "COMPOSE_PROJECT_NAME"
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

	composeVersion := cf.parseVersion(buf[0])
	if len(buf) > 1 {
		err = cf.validateVersions(buf)
		if err != nil {
			return nil, err
		}
	}

	loader := &ComposeLoader{}
	switch composeVersion {
	case "3":
		cf, err = loader.LoadVersion3(buf)
	default:
		cf, err = loader.LoadVersion2(buf)
	}

	if projectName != "" {
		cf.ProjectName = projectName
	} else {
		if cf.ProjectName == "" {
			if env, ok := os.LookupEnv(envComposeProjectName); ok {
				cf.ProjectName = env
			} else {
				cf.ProjectName = defaultProjectName
			}
		}
	}

	return
}

func (cf *ComposeFile) parseVersion(bytes []byte) string {
	var version composeVersion
	yaml.Unmarshal(bytes, &version)
	if version.Version == "" {
		return defaultVersion
	}
	return version.Version
}

func (cf *ComposeFile) validateVersions(in [][]byte) error {
	version := cf.parseVersion(in[0])
	for _, bytes := range in[1:] {
		if version != cf.parseVersion(bytes) {
			return fmt.Errorf("docker-compose version mismatch, want version: %s", version)
		}
	}
	return nil
}
