package command

import (
	"bytes"
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// StatsCommand is a Command implementation that is used to display usage
// statistics about an entire project.
type StatsCommand struct {
	UI       ui.UI
	Projects api.ProjectsGetter
	Apps     api.AppsGetMetricer
}

// Help should return long-form help text.
func (c *StatsCommand) Help() string {
	helpText := `
Usage: sloppy stats [OPTIONS] PROJECT

  Displays usage statistics of running instances(memory, traffic)

Options:
  -a, --all     Show all instances (default shows just running instances)

Examples:

  sloppy stats letschat
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *StatsCommand) Run(args []string) int {
	var all bool
	cmdFlags := newFlagSet("stats", flag.ContinueOnError)
	cmdFlags.BoolVar(&all, "a", false, "")
	cmdFlags.BoolVar(&all, "all", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy stats --help'.")
		return 1
	}

	if cmdFlags.NArg() < 1 {
		return c.UI.ErrorNotEnoughArgs("stats", "", 1)
	}

	if strings.Contains(cmdFlags.Arg(0), "/") {
		c.UI.Error(fmt.Sprintf("invalid project path \"%s\". \n", cmdFlags.Arg(0)))
		return 1
	}

	project, _, err := c.Projects.Get(cmdFlags.Arg(0))
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	stats, err := c.collect(project, all)
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	if len(stats) == 0 {
		c.UI.Output("No apps running")
		return 1
	}

	var buf bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(&buf, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "CONTAINER \t MEM / LIMIT \t MEM %% \t NET I/O Extern \t NET I/O Intern \t MAX VOLUME %% \t LAST UPDATE \n")

	var keys []string
	var latest api.Timestamp
	for _, stat := range stats {
		if stat.Time.After(latest.Time) {
			latest = stat.Time
		}
	}

	for k := range stats {
		diff := latest.Time.Sub(stats[k].Time.Time)
		if (diff < 5*time.Second && diff > -5*time.Second) || all {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "%s\n", stats[k])
	}

	w.Flush()

	c.UI.Output(buf.String())

	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *StatsCommand) Synopsis() string {
	return "Display metrics of a running app"
}

// Stat represents a container's stats
type stat struct {
	Service           string
	App               string
	Time              api.Timestamp
	ID                string // Container
	Memory            float64
	MemoryLimit       float64
	InternalNetworkRx float64
	InternalNetworkTx float64
	ExternalNetworkRx float64
	ExternalNetworkTx float64
	Volumes           int
	Volume            float64
}

func (s stat) String() string {
	return fmt.Sprintf("%s/%s-%s \t %s / %.f MiB \t %.1f%% \t %s / %s \t %s / %s \t %.1f%% \t %s",
		s.Service, s.App, s.ID[:6],
		humanByte(s.Memory), s.MemoryLimit,
		float64(s.Memory/(1<<20))/float64(s.MemoryLimit)*100,
		humanByte(s.ExternalNetworkRx), humanByte(s.ExternalNetworkTx),
		humanByte(s.InternalNetworkRx), humanByte(s.InternalNetworkTx),
		s.Volume,
		humanDuration(time.Now().Sub(s.Time.Time)),
	)
}

// Collect collects all statistics and merge them
func (c *StatsCommand) collect(project *api.Project, all bool) (map[string]*stat, error) {
	stats := make(map[string]*stat)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, service := range project.Services {
		wg.Add(len(service.Apps))
		for _, app := range service.Apps {
			statusCount := 0
			if !all {
				for i := range app.Status {
					if app.Status[i] == "running" {
						statusCount++
					}
				}
				if statusCount == 0 {
					wg.Done()
					continue
				}
			}

			go func(project *api.Project, service *api.Service, app *api.App) {
				defer wg.Done()
				metrics, _, err := c.Apps.GetMetrics(*project.Name, *service.ID, *app.ID)
				if err != nil {
					return
				}

				blueprint := stat{
					Service:     *service.ID,
					App:         *app.ID,
					MemoryLimit: float64(*app.Memory),
					Volumes:     len(app.Volumes),
				}

				// Sort keys in order to assign volume stats to the right container because go is accessing maps randomly.
				var keys []string
				for name := range metrics {
					keys = append(keys, name)
				}
				sort.Strings(keys)

				mutex.Lock()
				{
					for _, metric := range keys {
						switch metric {
						case "container_memory_usage_bytes":
							setMemory := func(s *stat, name string, p api.Point) {
								s.Memory = *p.Y
							}
							blueprint.SetMetrics(stats, metrics[metric], setMemory)
						case "container_volume_usage_percentage":
							setVolume := func(s *stat, name string, p api.Point) {
								// Only set if maximum
								if *p.Y > s.Volume {
									s.Volume = *p.Y
								}
							}
							blueprint.SetMetrics(stats, metrics[metric], setVolume)
						case "container_network_receive_bytes_per_second":
							setNetworkRx := func(s *stat, name string, p api.Point) {
								if i := strings.LastIndex(name, "eth0"); i != -1 {
									s.ExternalNetworkRx = *p.Y
								}
								if i := strings.LastIndex(name, "ethwe"); i != -1 {
									s.InternalNetworkRx = *p.Y
								}
							}
							blueprint.SetMetrics(stats, metrics[metric], setNetworkRx)
						case "container_network_transmit_bytes_per_second":
							setNetworkTx := func(s *stat, name string, p api.Point) {
								if i := strings.LastIndex(name, "eth0"); i != -1 {
									s.ExternalNetworkTx = *p.Y
								}
								if i := strings.LastIndex(name, "ethwe"); i != -1 {
									s.InternalNetworkTx = *p.Y
								}
							}
							blueprint.SetMetrics(stats, metrics[metric], setNetworkTx)
						}
					}
				}
				mutex.Unlock()

			}(project, service, app)
		}
	}
	wg.Wait()
	return stats, nil
}

func (s *stat) SetMetrics(stats map[string]*stat, series api.Series, set func(*stat, string, api.Point)) {
	for name, values := range series {
		if value := values[len(values)-1]; value.Y != nil {
			var id string
			i := strings.Index(name, ".") + 1

			if strings.Contains(name, "/") {
				for idstats := range stats {
					if s.Volumes > 0 && s.App == stats[idstats].App && s.Service == stats[idstats].Service {
						id = idstats
						set(stats[id], name, *value)
					}
				}
				continue
			}

			id = name[i : i+36]

			if stats[id] == nil {
				clone := *s
				clone.ID = id
				clone.Time = value.X
				stats[id] = &clone
			}
			set(stats[id], name, *value)
		}
	}
}

// HumanByte returns a human-readable size.
func humanByte(size float64) string {
	var abbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}

	i := 0
	for size >= 1024 {
		size = size / 1024
		i++
	}

	return fmt.Sprintf("%.3g %s", size, abbrs[i])
}

// HumanDuration returns a human-readable approximation of a duration
func humanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	} else if hours := int(d.Hours()); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("%d hours", hours)
	} else {
		return fmt.Sprintf("%d days", hours/24)
	}

}
