package converter

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

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

	for service, config := range cf.ServiceConfigs {
		if config.Build != nil {
			return nil, ErrBuildNotSupported
		}

		var uri *string
		if config.Domainname != "" {
			uri = &config.Domainname
		}

		app := &SloppyApp{
			App: &sloppy.App{
				Image:   &config.Image,
				Volumes: sf.convertVolumes(config.Volumes),
			},
		}

		// Assign possible empty values in extra steps to hide empty object from output
		// Commands (string or list)
		switch config.Command.(type) {
		case string:
			c := config.Command.(string)
			app.App.Command = &c
		case []interface{}:
			c := config.Command.([]interface{})
			if len(c) > 0 {
				app.App.Command = sf.convertCommand(c)
			}
		}

		// Domain
		if uri != nil {
			app.App.Domain = &sloppy.Domain{URI: uri}
			app.Domain = uri
		}

		if envList, ok := config.Environment.([]interface{}); ok {
			app.App.EnvVars = make(map[string]string)
			for _, e := range envList {
				env := e.(string)
				split := strings.Split(env, "=")
				k, v := split[0], split[1]
				app.App.EnvVars[k] = v
				app.Env = append(app.Env, map[string]string{k: v})
			}
		} else if envmap, ok := config.Environment.(map[string]interface{}); len(envmap) > 0 && ok {
			app.App.EnvVars = make(map[string]string)
			for v, val := range envmap {
				app.App.EnvVars[v] = val.(string)
			}
			for k, v := range app.App.EnvVars {
				app.Env = append(app.Env, map[string]string{k: v})
			}
		}

		// Logging
		if config.Logging != nil && config.Logging.Driver != "" && config.Logging.Options != nil {
			logOpts, ok := config.Logging.Options.(map[string]interface{})
			if ok && len(logOpts) > 0 {
				optsMap := make(map[string]string)
				for k, v := range logOpts {
					optsMap[k] = v.(string)
				}
				app.App.Logging = &sloppy.Logging{
					Driver:  &config.Logging.Driver,
					Options: optsMap,
				}
			}
		}

		// Port
		if len(config.Ports) > 0 {
			var ports []string
			for _, entry := range config.Ports {
				switch entry.(type) { // number, string, obj
				case string:
					ports = append(ports, entry.(string))
				case int:
					p := strconv.Itoa(entry.(int))
					ports = append(ports, p)
				}
			}
			portMappings, err := sf.convertPorts(ports)
			if err != nil {
				return nil, err
			}

			// In yml format just one port is supported, use the first one.
			// And don't set app.App.PortMappings.
			app.Port = portMappings[0].Port
		}

		if config.Deploy != nil {
			if config.Deploy.Replicas > 0 {
				app.Instances = &config.Deploy.Replicas
			}
			if config.Deploy.Resources != nil &&
				config.Deploy.Resources.Limits != nil {
				var err error
				app.Memory, err = sf.convertMemoryResource(config.Deploy.Resources.Limits.Memory)
				if err != nil {
					return nil, err
				}
			}
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

func (sf *SloppyFile) convertCommand(cmd []interface{}) *string {
	var str string
	for i, s := range cmd {
		str += s.(string)
		if i < len(cmd)-1 {
			str += " "
		}
	}
	return &str
}

var resourceMemRegex = regexp.MustCompile(`^(\d+)([bkmgBKMG]){1}$`)

func (sf *SloppyFile) convertMemoryResource(res string) (*int, error) {
	const lowestMem float64 = 64 // lowest mem size sloppy.io supports
	match := resourceMemRegex.FindStringSubmatch(res)
	if len(match) == 3 {
		memResource, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, err
		}
		unit := strings.ToLower(match[2])
		var mem int
		switch unit {
		case "b":
			mem = int(math.Max(float64(memResource/1024/1024), lowestMem))
		case "k":
			mem = int(math.Max(float64(memResource/1024), lowestMem))
		case "m":
			mem = memResource
		case "g":
			mem = memResource * 1024
		default:
			return nil, fmt.Errorf("convert resources: unsupported memory format: %q", res)
		}
		return &mem, nil
	}
	return nil, nil
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

func (sf *SloppyFile) toVolumes(str string) map[string]string {
	const sep = ":"
	out := make(map[string]string)
	if strings.Index(str, sep) > 0 {
		parts := strings.Split(str, sep)
		out["source"], out["target"] = parts[0], parts[1]
		if strings.Index(parts[1], sep) > 0 {
			perms := strings.Split(parts[1], sep)
			out["accessMode"] = perms[1]
		}
	} else {
		out["target"] = str
	}
	return out
}

func (sf *SloppyFile) convertVolumes(volumes []interface{}) (v []*sloppy.Volume) {
	if volumes == nil {
		return
	}
	for _, volume := range volumes {
		if vstring, ok := volume.(string); ok {
			dest := sf.toVolumes(vstring)["target"]
			v = append(v, &sloppy.Volume{
				Path: &dest,
			})
		} else if vmap, ok := volume.(map[string]interface{}); ok {
			dest := vmap["target"].(string)
			v = append(v, &sloppy.Volume{
				Path: &dest,
			})
		}
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
