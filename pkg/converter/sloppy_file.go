package converter

import (
	"strconv"
	"strings"

	compose "github.com/docker/libcompose/yaml"
	sloppy "github.com/sloppyio/cli/src/api"
)

const (
	// defaults
	instanceCount  = 1
	instanceMemory = 512
	volumeSize     = "8GB"
)

type SloppyApps map[string]*sloppy.App

type SloppyFile struct {
	Version  string                `yaml:"version,omitempty"`
	Project  string                `yaml:"project,omitempty"`
	Services map[string]SloppyApps `yaml:"services,omitempty"`
}

// Map docker-compose.yml to sloppy yml and return representation
func NewSloppyFile(cf *ComposeFile) (*SloppyFile, error) {
	sf := &SloppyFile{
		Version:  "v1",
		Project:  cf.ProjectName,
		Services: map[string]SloppyApps{"apps": make(SloppyApps)},
	}

	for service, config := range cf.ServiceConfigs.All() {
		m, i := instanceMemory, instanceCount
		app := &sloppy.App{
			Memory:    &m,
			Instances: &i,
			Image:     &config.Image,
			EnvVars:   config.Environment.ToMap(),
			Volumes:   sf.convertVolumes(config.Volumes),
		}

		// assign command
		if len(config.Command) > 0 {
			app.Command = sf.convertCommand(config.Command)
		}

		// assign ports
		if len(config.Ports) > 0 {
			portMappings, err := sf.convertPorts(config.Ports)
			if err != nil {
				return nil, err
			}
			app.PortMappings = portMappings
		}

		// TODO implement service to compose-file mapping
		// sloppy naming:
		//  []   = service
		//  [][] = app
		sf.Services["apps"][service] = app
	}
	return sf, nil
}

func (sf *SloppyFile) convertCommand(cmd compose.Command) *string {
	var str string
	for i, s := range cmd {
		str += s
		if i < len(cmd)-1 {
			str += " "
		}
	}
	return &str
}

func (sf *SloppyFile) convertPorts(ports []string) (pm []*sloppy.PortMap, err error) {
	const sep = ":"
	for _, portMap := range ports {
		var port int
		if strings.Index(portMap, sep) > -1 {
			port, err = strconv.Atoi(strings.Split(portMap, sep)[1])
		} else {
			port, err = strconv.Atoi(portMap)
		}
		if err != nil {
			return nil, err
		}
		pm = append(pm, &sloppy.PortMap{Port: &port})
	}

	return
}

func (sf *SloppyFile) convertVolumes(volumes *compose.Volumes) (v []*sloppy.Volume) {
	defaultSize := volumeSize
	if volumes == nil {
		return
	}
	for _, volume := range volumes.Volumes {
		v = append(v, &sloppy.Volume{
			Path: &volume.Destination,
			Size: &defaultSize,
		})
	}
	return
}
