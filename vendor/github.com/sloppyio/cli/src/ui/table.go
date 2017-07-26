package ui

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/sloppyio/cli/src/api"
)

func tableStart(w io.Writer, services []*api.Service) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"SERVICE", "# APPS", "TOTAL MEMORY"})

	for _, service := range services {
		totalMemory := 0
		for _, app := range service.Apps {
			totalMemory += *app.Memory * *app.Instances

		}

		table.Append([]string{
			*service.ID,
			fmt.Sprintf("%d", len(service.Apps)),
			fmt.Sprintf("%d MiB", totalMemory),
		})
	}

	table.Render()
}

// TableShow prints tables depending on type of v.
func tableShow(w io.Writer, v interface{}) {
	switch t := v.(type) {
	case []api.Project:
		tableProjects(w, t)
	case []*api.Service:
		tableServices(w, t)
	case []*api.App:
		tableApps(w, t)
	case *api.App:
		listApp(w, t)
	}
}

func tableProjects(w io.Writer, projects []api.Project) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"PROJECT", "# SERVICES", "# APPS", "TOTAL MEMORY"})

	for _, project := range projects {
		countApp := 0
		totalMemory := 0

		for _, service := range project.Services {
			countApp += len(service.Apps)
			for _, app := range service.Apps {
				totalMemory += *app.Memory * *app.Instances
			}
		}

		table.Append([]string{
			*project.Name,
			fmt.Sprintf("%d", len(project.Services)),
			fmt.Sprintf("%d", countApp),
			fmt.Sprintf("%d MiB", totalMemory),
		})
	}

	table.Render()
}

func tableServices(w io.Writer, services []*api.Service) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"SERVICE", "# APPS", "STATUS", "TOTAL MEMORY"})

	for _, service := range services {
		totalMemory := 0
		runningContainer := 0
		totalContainter := 0
		for _, app := range service.Apps {
			totalMemory += *app.Memory * *app.Instances
			totalContainter += *app.Instances
			runningContainer += countMatch(app.Status, "running")
		}

		table.Append([]string{
			*service.ID,
			fmt.Sprintf("%d", len(service.Apps)),
			fmt.Sprintf("%d / %d", runningContainer, totalContainter),
			fmt.Sprintf("%d MiB", totalMemory),
		})
	}

	table.Render()
}

func tableApps(w io.Writer, apps []*api.App) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"APP", "IMAGE", "COMMAND", "TOTAL MEMORY", "STATUS"})

	for _, app := range apps {
		table.Append([]string{
			*app.ID,
			*app.Image,
			fmt.Sprintf("%s", placeholder(app.Command, "-")),
			fmt.Sprintf("%d MiB", *app.Memory**app.Instances),
			fmt.Sprintf("%d / %d", countMatch(app.Status, "running"), *app.Instances),
		})
	}

	table.Render()
}

func tableApp(w io.Writer, app *api.App) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"APP", "IMAGE", "VERSION", "DOMAIN", "COMMAND", "DEPENDENCIES", "PORTS", "ENV", "MEMORY", "STATUS"})

	var envs []string
	for key, value := range app.EnvVars {
		envs = append(envs, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	var ports []string
	for _, value := range app.PortMappings {
		ports = append(ports, strconv.Itoa(*value.Port))
	}

	table.Append([]string{
		*app.ID,
		*app.Image,
		*app.Version,
		fmt.Sprintf("%s", *app.Domain.URI),
		fmt.Sprintf("%s", placeholder(app.Command, "-")),
		fmt.Sprintf("%s", strings.Join(app.Dependencies, ";")),
		fmt.Sprintf("%s", strings.Join(ports, ",")),
		strings.Join(envs, "; "),
		fmt.Sprintf("%d MiB", *app.Memory**app.Instances),
		fmt.Sprintf("%d / %d", countMatch(app.Status, "running"), *app.Instances),
	})
	table.Render()
}

func listApp(w io.Writer, app *api.App) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Application: \t %s\n", *app.ID)
	fmt.Fprintf(&buf, "Version: \t %s\n", *app.Version)
	if *app.Instances < 2 {
		fmt.Fprintf(&buf, "Memory:\t\t %d MiB\n", *app.Memory)
	} else {
		fmt.Fprintf(&buf, "Memory:\t\t %d x %d MiB\n", *app.Instances, *app.Memory)
	}
	fmt.Fprintf(&buf, "Instances:\t %d / %d\n", countMatch(app.Status, "running"), *app.Instances)
	var domain = "-"
	if app.Domain != nil {
		domain = placeholder(app.Domain.URI, "-")
	}
	fmt.Fprintf(&buf, "Domain:\t\t %s\n", domain)
	fmt.Fprintf(&buf, "Image:\t\t %s\n", *app.Image)
	fmt.Fprintf(&buf, "Command:\t %s\n", placeholder(app.Command, "-"))

	if len(app.Volumes) == 0 {
		fmt.Fprintf(&buf, "Volumes:\t -\n")
	} else {
		fmt.Fprintf(&buf, "Volumes:")
	}

	for i, volume := range app.Volumes {
		ident := "\t"
		if i != 0 {
			ident = "\t\t"
		}
		fmt.Fprintf(&buf, ident+" '%s' %s\n", *volume.Path, *volume.Size)
	}

	if len(app.PortMappings) == 0 {
		fmt.Fprintf(&buf, "Ports:\t\t -\n")
	} else {
		fmt.Fprintf(&buf, "Ports:")
	}
	for _, value := range app.PortMappings {
		fmt.Fprintf(&buf, "\t\t %d\n", *value.Port)
	}

	if app.Logging != nil {
		fmt.Fprintf(&buf, "Logging:\n")
		if app.Logging.Driver != nil {
			fmt.Fprintf(&buf, "  Driver:\t %s\n", *app.Logging.Driver)
		} else {
			fmt.Fprintf(&buf, "  Driver:\t -\n")
		}
		first := true
		for k, v := range app.Logging.Options {
			if first {
				fmt.Fprintf(&buf, "  Options:\t %s=%q\n", k, v)
				first = false
			} else {
				fmt.Fprintf(&buf, "\t\t %s=%q\n", k, v)
			}
		}
	}

	listing(&buf, "Dependencies", "\t", "\t\t", app.Dependencies)

	if len(app.EnvVars) == 0 {
		fmt.Fprintf(&buf, "Environments:\t -\n")
	} else {
		fmt.Fprintf(&buf, "Environments:")
	}
	var first = true
	for key, value := range app.EnvVars {
		t := "\t"
		if !first {
			t = "\t\t"
		}
		first = false
		fmt.Fprintf(&buf, t+" %s=\"%s\"\n", key, value)
	}

	listing(&buf, "Versions", "\t", "\t\t", app.Versions)

	io.Copy(w, &buf)
}

// countMatch counts values in the given slice which matches the pattern.
func countMatch(slice []string, pattern string) int {
	n := 0

	for _, value := range slice {
		if value == pattern {
			n++
		}
	}

	return n
}

// Null pointer returns a placeholder
func placeholder(v *string, placeholder string) string {
	if v == nil {
		return placeholder
	}
	return *v
}

func listing(w io.Writer, name, indent, indent2 string, values []string) {
	if len(values) == 0 {
		if len(name) < 8 {
			indent = indent2
		}
		fmt.Fprintf(w, "%s:%s -\n", name, indent)
		return
	}

	for i, value := range values {
		if i == 0 {
			if len(name) < 8 {
				indent = indent2
			}
			fmt.Fprintf(w, "%s:%s %s\n", name, indent, value)
			continue
		}
		fmt.Fprintf(w, "%s %s\n", indent2, value)
	}
}
