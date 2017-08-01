package converter

import (
	"fmt"
	"regexp"
	"strings"

	sloppy "github.com/sloppyio/cli/src/api"
)

const (
	fqdnTemplate    = "%s.%s.%s"                   // app.service.project
	hostPortPattern = "([a-z]+[a-z0-9_-]?):[0-9]+" // sloppy appName conform
)

var hostPortRegex *regexp.Regexp = regexp.MustCompile(hostPortPattern)

type DependencyError struct {
	errStr string
}

type Linker struct {
	links []*link
}

type link struct {
	app     *SloppyApp
	fqdn    string
	ports   []*sloppy.PortMap
	appName string
}

func newDependencyError(msg string, args ...string) *DependencyError {
	return &DependencyError{fmt.Sprintf(msg, args)}
}

func (d *DependencyError) Error() string {
	return d.errStr
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
	l.buildLinks(sf)

	// resolve possible connections
	for _, link := range l.links {
		for key, val := range link.app.App.EnvVars {
			app := link.app.App
			if strings.Contains(key, "HOST") ||
				strings.Index(val, ":") != -1 {
				matches := hostPortRegex.FindStringSubmatch(val)
				if matches == nil {
					continue
				}
				match := matches[1]
				targetLink := l.GetByApp(match)

				if targetLink == nil {
					fmt.Printf("Couldn't find %q as linkable app. Assuming %q is an external service.\n", match, val)
					continue
				}

				targetVar := strings.Replace(
					app.EnvVars[key],
					match,
					targetLink.fqdn,
					1,
				)
				app.EnvVars[key] = targetVar

				// also consider special sloppy Env field
				for i, s := range link.app.Env {
					if s == strings.Join([]string{key, val}, "=") {
						link.app.Env[i] = strings.Join([]string{key, targetVar}, "=")
						break
					}
				}

				app.Dependencies = l.appendDependency(link.app, targetLink.fqdn)
			}

			// also considering DependsOn and Links from compose
			for serviceName, conf := range cf.ServiceConfigs.All() {
				if serviceName != link.appName {
					continue
				}
				if len(conf.DependsOn) > 0 {
					for _, dep := range conf.DependsOn {
						t := l.GetByApp(dep)
						if t == nil {
							return newDependencyError(`Couldn't find related service %q declared in "depends_on"`, dep)
						}
						app.Dependencies = l.appendDependency(link.app, t.fqdn)
					}
				}

				if len(conf.Links) > 0 {
					for _, val := range conf.Links {
						t := l.GetByApp(val)
						if t == nil {
							return newDependencyError(`Couldn't find related service %q declared in "links"`, val)
						}
						app.Dependencies = l.appendDependency(link.app, t.fqdn)
					}
				}
			}
		}
	}

	sf.sortFields()

	return nil
}

func (l *Linker) appendDependency(app *SloppyApp, fqdn string) []string {
	s := l.formatDependency(fqdn)
	if len(app.App.Dependencies) > 0 {
		for _, dep := range app.App.Dependencies {
			if s == dep {
				return app.App.Dependencies
			}
		}
	}
	return append(
		app.App.Dependencies,
		s,
	)
}

func (l *Linker) buildLinks(sf *SloppyFile) {
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
}

func (l *Linker) formatDependency(in string) (out string) {
	parts := strings.Split(in, ".")
	out = ".."
	for i := len(parts) - 1; i > 0; i-- {
		out += fmt.Sprintf("/%s", parts[i-1])
	}
	return
}
