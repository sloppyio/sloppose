package converter

import (
	"fmt"
	"regexp"
	"strings"

	sloppy "github.com/sloppyio/cli/src/api"
)

const (
	fqdnTemplate    = "%s.%s.%s"                       // app.service.project
	hostPortPattern = `([a-z]+[a-z0-9._-]*)(:[0-9]+)?` // sloppy appName conform
	schemePattern   = `^(\w+)(:\/\/)+`
)

var hostPortRegex *regexp.Regexp = regexp.MustCompile(hostPortPattern)
var schemeRegex *regexp.Regexp = regexp.MustCompile(schemePattern)

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
			match := l.FindServiceString(key, val)
			if match == "" {
				continue
			}

			targetLink := l.GetByApp(match)

			if targetLink == nil {
				fmt.Printf("Couldn't find %q as linkable app. Assuming %q is an external service.\n", match, val)
				continue
			}

			var targetVar string
			schemeIdx := strings.Index(app.EnvVars[key], "://")
			if schemeIdx != -1 {
				targetVar = app.EnvVars[key][:schemeIdx] + strings.Replace(
					app.EnvVars[key][schemeIdx:],
					match,
					targetLink.fqdn,
					1,
				)
			} else {
				targetVar = strings.Replace(
					app.EnvVars[key],
					match,
					targetLink.fqdn,
					1,
				)
			}

			app.EnvVars[key] = targetVar

			// also consider special sloppy Env field
			for i, kv := range link.app.Env {
				if _, ok := kv[key]; ok {
					link.app.Env[i][key] = targetVar
					break
				}
			}

			app.Dependencies = l.appendDependency(link.app, targetLink.fqdn)
		}

		// also considering DependsOn from compose
		if conf, ok := cf.ServiceConfigs.Get(link.appName); ok {
			if len(conf.DependsOn) > 0 {
				for _, dep := range conf.DependsOn {
					t := l.GetByApp(dep)
					if t == nil {
						return newDependencyError(`Couldn't find related service %q declared in "depends_on"`, dep)
					}
					link.app.Dependencies = l.appendDependency(link.app, t.fqdn)
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

// Searches for services in environment variable values.
// Primary match would be a <host:port> one.
// To also support service linking without a port the
// environment key name requires to contain the `HOST` string.
func (l *Linker) FindServiceString(key string, val string) string {
	if strings.Index(val, ":") != -1 ||
		strings.Contains(key, "HOST") {
		matches := hostPortRegex.FindAllStringSubmatch(val, -1)
		for _, subMatch := range matches {
			schemeMatches := schemeRegex.FindStringSubmatch(val)
			// skipping schemes as service
			if schemeMatches != nil && subMatch[1] == schemeMatches[1] {
				continue
			}

			return subMatch[1]
		}
	}
	return ""
}

func (l *Linker) formatDependency(in string) (out string) {
	parts := strings.Split(in, ".")
	out = ".."
	for i := len(parts) - 1; i > 0; i-- {
		out += fmt.Sprintf("/%s", parts[i-1])
	}
	return
}
