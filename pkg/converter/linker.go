package converter

import (
	"fmt"
	"regexp"
	"strings"

	sloppy "github.com/sloppyio/cli/src/api"
)

const (
	fqdnTemplate     = "%s.%s.%s" // app.service.project
	lowercasePattern = "^[a-z]*"
)

var lowercase *regexp.Regexp = regexp.MustCompile(lowercasePattern)

type Linker struct {
	links []*link
}

type link struct {
	app     *sloppy.App
	fqdn    string
	ports   []*sloppy.PortMap
	appName string
}

func (l *Linker) GetByApp(name string) *link {
	for _, link := range l.links {
		if strings.HasPrefix(link.fqdn, name) {
			return link
		}
	}
	return nil
}

func (l *Linker) Resolve(cf *ComposeFile, sf *SloppyFile) error {
	// build possible links
	for serviceName, apps := range sf.Services {
		for appName, app := range apps {
			l.links = append(
				l.links, &link{
					app:     app,
					fqdn:    fmt.Sprintf(fqdnTemplate, appName, serviceName, sf.Project),
					ports:   app.PortMappings,
					appName: appName,
				},
			)
		}
	}

	appendDependency := func(app *sloppy.App, fqdn string) {
		s := l.formatDependency(fqdn)
		if len(app.Dependencies) > 0 {
			for _, dep := range app.Dependencies {
				if s == dep {
					return
				}
			}
		}
		app.Dependencies = append(
			app.Dependencies,
			s,
		)
	}

	// resolve possible connections
	for _, link := range l.links {
		for key, val := range link.app.EnvVars {
			if strings.Contains(key, "HOST") ||
				strings.Index(val, ":") != -1 {
				match := lowercase.FindString(val)
				targetLink := l.GetByApp(match)
				//fmt.Println("Replacing:", key, val, match, targetLink.fqdn)
				if targetLink == nil {
					return fmt.Errorf("Couldn't find app %q", match)
				}
				link.app.EnvVars[key] = strings.Replace(
					link.app.EnvVars[key],
					match,
					targetLink.fqdn,
					1,
				)

				appendDependency(link.app, targetLink.fqdn)
			}

			// also considering DependsOn and Links from compose
			for serviceName, conf := range cf.ServiceConfigs.All() {
				if serviceName != link.appName {
					continue
				}
				if len(conf.DependsOn) > 0 {
					for _, dep := range conf.DependsOn {
						t := l.GetByApp(dep)
						if t != nil {
							appendDependency(link.app, t.fqdn)
						}
					}
				}

				if len(conf.Links) > 0 {
					for _, val := range conf.Links {
						t := l.GetByApp(val)
						if t != nil {
							appendDependency(link.app, t.fqdn)
						}
					}
				}
			}
		}
	}

	return nil
}

func (l *Linker) formatDependency(in string) (out string) {
	parts := strings.Split(in, ".")
	out = ".."
	for i := len(parts) - 1; i > 0; i-- {
		out += fmt.Sprintf("/%s", parts[i-1])
	}
	return
}
