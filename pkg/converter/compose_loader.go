package converter

import (
	"fmt"
	"os"
	"strings"

	v3 "github.com/aanand/compose-file/loader"
	v3types "github.com/aanand/compose-file/types"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
)

type ComposeLoader struct{}

func (cl *ComposeLoader) LoadVersion3(buf [][]byte) (cf *ComposeFile, err error) {
	var v3ConfFiles []v3types.ConfigFile
	for _, bytes := range buf {
		dict, err := v3.ParseYAML(bytes)
		if err != nil {
			return nil, err
		}
		v3ConfFiles = append(v3ConfFiles, v3types.ConfigFile{Config: dict})
	}

	cwd, _ := os.Getwd()
	conf, err := v3.Load(v3types.ConfigDetails{
		WorkingDir:  cwd,
		ConfigFiles: v3ConfFiles,
	})
	if err != nil {
		return nil, err
	}
	return cl.ConvertToV2(conf)
}

func (cl *ComposeLoader) LoadVersion2(buf [][]byte) (*ComposeFile, error) {
	cf := &ComposeFile{ServiceConfigs: config.NewServiceConfigs()}
	ctx := &project.Context{
		ComposeBytes: buf,
	}

	p := project.NewProject(ctx, nil, nil)
	err := p.Parse()
	if err != nil {
		return nil, err
	}

	// available after project.Parse() if not previously set
	cf.ProjectName = ctx.ProjectName

	cfg := *p.ServiceConfigs
	cf.ServiceConfigs = &cfg
	return cf, nil
}

// Converts a compose version 3 format to version 2.
// Since sloppyio supports just a small subset of compose commands this is the current approach.
func (cl *ComposeLoader) ConvertToV2(v3conf *v3types.Config) (*ComposeFile, error) {
	cf := &ComposeFile{ServiceConfigs: config.NewServiceConfigs()}
	toEqualSlice := func(in map[string]string) (out yaml.MaporEqualSlice) {
		for k, v := range in {
			out = append(out, fmt.Sprintf("%s=%s", k, v))
		}
		return
	}

	toVolumes := func(in []string) *yaml.Volumes {
		const sep = ":"
		out := &yaml.Volumes{}
		for _, str := range in {
			volume := &yaml.Volume{}
			if strings.Index(str, sep) > 0 {
				parts := strings.Split(str, sep)
				volume.Source, volume.Destination = parts[0], parts[1]
				if strings.Index(parts[1], sep) > 0 {
					perms := strings.Split(parts[1], sep)
					volume.AccessMode = perms[1]
				}
			} else {
				volume.Destination = str
			}
			out.Volumes = append(out.Volumes, volume)
		}
		return out
	}

	for _, service := range v3conf.Services {
		v2conf := &config.ServiceConfig{
			DependsOn:   service.DependsOn,
			DomainName:  service.DomainName,
			Entrypoint:  service.Entrypoint,
			Environment: toEqualSlice(service.Environment),
			Expose:      service.Expose,
			Image:       service.Image,
			Labels:      service.Labels,
			Links:       service.Links,
			Logging: config.Log{
				Driver:  service.Logging.Driver,
				Options: service.Logging.Options,
			},
			//MemLimit: service.Deploy.Resources.Limits.MemoryBytes, TODO convert to xxxMB
			Ports:      service.Ports,
			WorkingDir: service.WorkingDir,
			Volumes:    toVolumes(service.Volumes),
		}
		cf.ServiceConfigs.Add(service.Name, v2conf)
	}
	return cf, nil
}
