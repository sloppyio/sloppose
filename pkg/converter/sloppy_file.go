package converter

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	compose "github.com/docker/libcompose/yaml"
	sloppy "github.com/sloppyio/cli/pkg/api"
)

var (
	ErrBuildNotSupported = errors.New("the build property is not supported, please specify an image instead")
)

type SloppyApps map[string]*SloppyApp

type SloppyEnvSlice []map[string]string

// Special intermediate type to fix the inconsistencies between
// the yml and json format representation for the sloppy.App struct.
type SloppyApp struct {
	*sloppy.App
	Domain *string        `json:"domain,omitempty"`
	Env    SloppyEnvSlice `json:"env,omitempty"`
	Port   *int           `json:"port,omitempty"`

	// hide conflicting fields from sloppy.App during serialization
	EnvVars      map[string]string `json:"-"`
	PortMappings []*sloppy.PortMap `json:"-"`
}

type SloppyFile struct {
	Version  string                `json:"version,omitempty"`
	Project  string                `json:"project,omitempty"`
	Services map[string]SloppyApps `json:"services,omitempty"`
}

func (p SloppyEnvSlice) Len() int { return len(p) }
func (p SloppyEnvSlice) Less(i, j int) bool {
	for k := range p[i] {
		for ok := range p[j] {
			return k < ok
		}
	}
	return false
}
func (p SloppyEnvSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Map docker-compose.yml to sloppy yml and return representation
func NewSloppyFile(cf *ComposeFile) (*SloppyFile, error) {
	sf := &SloppyFile{
		Version:  "v1",
		Project:  cf.ProjectName,
		Services: map[string]SloppyApps{"apps": make(SloppyApps)},
	}

	for service, config := range cf.ServiceConfigs.All() {
		if config.Build.Context != "" {
			return nil, ErrBuildNotSupported
		}

		var uri *string
		if config.DomainName != "" {
			uri = &config.DomainName
		}

		app := &SloppyApp{
			App: &sloppy.App{
				Image:   &config.Image,
				Volumes: sf.convertVolumes(config.Volumes),
			},
		}

		// Assign possible empty values in extra steps to hide empty object from output
		// Commands
		if len(config.Command) > 0 {
			app.App.Command = sf.convertCommand(config.Command)
		}

		// Domain
		if uri != nil {
			app.App.Domain = &sloppy.Domain{URI: uri}
			app.Domain = uri
		}

		// Environment
		if len(config.Environment) > 0 {
			app.App.EnvVars = config.Environment.ToMap()
			for k, v := range app.App.EnvVars {
				app.Env = append(app.Env, map[string]string{k: v})
			}
		}

		// Logging
		if config.Logging.Driver != "" && len(config.Logging.Options) > 0 {
			app.App.Logging = &sloppy.Logging{
				Driver:  &config.Logging.Driver,
				Options: config.Logging.Options,
			}
		}

		// Port
		if len(config.Ports) > 0 {
			portMappings, err := sf.convertPorts(config.Ports)
			if err != nil {
				return nil, err
			}

			// In yml format just one port is supported, use the first one.
			// And don't set app.App.PortMappings.
			app.Port = portMappings[0].Port
		}

		// TODO implement service to compose-file mapping
		// Possible option to map multiple compose-files to own sloppy services
		// instead of the current default "apps"

		// sloppy naming:
		//  []   = service
		//  [][] = app
		sf.Services["apps"][service] = app
	}

	sf.sortFields()

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
			return nil, fmt.Errorf("port ranges are not supported: %q", portMap)
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
	if volumes == nil {
		return
	}
	for _, volume := range volumes.Volumes {
		v = append(v, &sloppy.Volume{
			Path: &volume.Destination,
		})
	}
	return
}

// Sorting the converted string slices ensures that
// the serialized output is always the same.
func (sf *SloppyFile) sortFields() {
	for _, services := range sf.Services {
		for _, app := range services {
			sort.Sort(app.Env)
			sort.Strings(app.App.Dependencies)
		}
	}
}
