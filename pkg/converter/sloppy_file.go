package converter

import (
	"fmt"
	"strconv"
	"strings"

	compose "github.com/docker/libcompose/yaml"
	sloppy "github.com/sloppyio/cli/src/api"
)

const (
	// defaults
	DomainUri      = "$URI"
	InstanceCount  = 1
	InstanceMemory = 512
	VolumeSize     = "8GB"
)

type SloppyApps map[string]*SloppyApp

// Special intermediate type to fix the inconsistencies between
// the yml and json format representation for the sloppy.App struct.
type SloppyApp struct {
	*sloppy.App
	Domain string                  `json:"domain,omitempty"`
	Env    compose.MaporEqualSlice `json:"env,omitempty"`
	Port   *int                    `json:"port,omitempty"`

	// hide conflicting fields from sloppy.App
	EnvVars      map[string]string `json:"-"`
	PortMappings []*sloppy.PortMap `json:"-"`
}

type SloppyFile struct {
	Version  string                `json:"version,omitempty"`
	Project  string                `json:"project,omitempty"`
	Services map[string]SloppyApps `json:"services,omitempty"`
}

// Map docker-compose.yml to sloppy yml and return representation
func NewSloppyFile(cf *ComposeFile) (*SloppyFile, error) {
	sf := &SloppyFile{
		Version:  "v1",
		Project:  cf.ProjectName,
		Services: map[string]SloppyApps{"apps": make(SloppyApps)},
	}

	for service, config := range cf.ServiceConfigs.All() {
		m, i, uri := InstanceMemory, InstanceCount, DomainUri
		if config.DomainName != "" {
			uri = config.DomainName
		}
		app := &SloppyApp{
			App: &sloppy.App{
				Domain:    &sloppy.Domain{URI: &uri},
				Image:     &config.Image,
				Instances: &i,
				Memory:    &m,
				Volumes:   sf.convertVolumes(config.Volumes),
			},
			Domain: uri,
		}

		// assign command
		if len(config.Command) > 0 {
			app.Command = sf.convertCommand(config.Command)
		}

		if len(config.Environment) > 0 {
			app.App.EnvVars = config.Environment.ToMap()
			app.Env = config.Environment
		}

		// assign ports
		if len(config.Ports) > 0 {
			portMappings, err := sf.convertPorts(config.Ports)
			if err != nil {
				return nil, err
			}
			app.App.PortMappings = portMappings

			// In yml format just one port is supported, use the first one.
			app.Port = portMappings[0].Port
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
		if strings.Index(portMap, "-") > -1 {
			return nil, fmt.Errorf("Port ranges are not supported: %q", portMap)
		}
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
	defaultSize := VolumeSize
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
