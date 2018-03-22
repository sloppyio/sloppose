package converter

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/sloppyio/sloppose/pkg/config"
)

const (
	DefaultProjectName    = "sloppyio"
	EnvComposeProjectName = "COMPOSE_PROJECT_NAME"
)

var (
	ErrFileRequired   = errors.New("at least one read file is required")
	ErrMissingVersion = errors.New("missing version declaration in compose file")
	ErrComposeVersion = errors.New("given compose version is not supported")
)

type ComposeFile struct {
	ProjectName    string
	ServiceConfigs map[string]*config.Service
}

type composeVersion struct {
	Version string `yaml:"version"`
}

func NewComposeFile(buf []byte, projectName string) (cf *ComposeFile, err error) {
	if len(buf) == 0 {
		return nil, ErrFileRequired
	}

	composeVersion, err := cf.parseVersion(buf)
	if err != nil {
		return nil, err
	}

	loader := &ComposeLoader{}
	switch strings.Split(composeVersion, ".")[0] {
	case "3":
		cf, err = loader.LoadVersion3(buf)
	case "2":
		err = ErrComposeVersion
	default:
		err = ErrMissingVersion
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
	cf.ProjectName = strings.ToLower(cf.ProjectName) // TODO remove _- ?

	err = cf.loadEnvFile()
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

func (cf *ComposeFile) loadEnvFile() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, service := range cf.ServiceConfigs {
		if service.EnvFile == nil {
			continue
		}
		var files []string
		switch service.EnvFile.(type) { // string or list
		case string:
			files = append(files, service.EnvFile.(string))
		case []interface{}:
			envs := service.EnvFile.([]interface{})
			for _, e := range envs {
				files = append(files, e.(string))
			}
		}

		var vars []interface{}
		for i := len(files) - 1; i >= 0; i-- {
			envFile := path.Join(cwd, files[i])
			content, err := ioutil.ReadFile(envFile)
			if err != nil {
				return err
			}
			scanner := bufio.NewScanner(bytes.NewBuffer(content))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				if len(line) > 0 && !strings.HasPrefix(line, "#") {
					key := strings.SplitAfter(line, "=")[0]

					found := false
					for _, v := range vars {
						if strings.HasPrefix(v.(string), key) {
							found = true
							break
						}
					}

					if !found {
						vars = append(vars, line)
					}
				}
			}

			if scanner.Err() != nil {
				scanner.Err()
			}
		}
		service.Environment = vars

		// delete
		service.EnvFile = nil
	}
	return nil
}

func (cf *ComposeFile) parseVersion(bytes []byte) (string, error) {
	var version composeVersion
	err := yaml.Unmarshal(bytes, &version)
	if err != nil {
		return "", err
	}
	return version.Version, nil
}
