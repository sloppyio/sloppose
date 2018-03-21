package converter

import (
	"github.com/ghodss/yaml"

	"github.com/sloppyio/sloppose/pkg/config"
)

type ComposeLoader struct{}

func (cl *ComposeLoader) LoadVersion3(buf []byte) (*ComposeFile, error) {
	composeFile := &config.DockerComposeV3{}
	err := yaml.Unmarshal(buf, composeFile)
	if err != nil {
		return nil, err
	}
	return &ComposeFile{
		ServiceConfigs: composeFile.Services,
	}, nil
}
