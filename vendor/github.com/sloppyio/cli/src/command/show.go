package command

import (
	"flag"
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// ShowCommand is a Command implementation that is used to display detailed
// information on either all your projects; or just one specific project,
// service or application.
type ShowCommand struct {
	UI       ui.UI
	Projects api.ProjectsGetLister
	Services api.ServicesGetter
	Apps     api.AppsGetter
}

// Help should return long-form help text.
func (c *ShowCommand) Help() string {
	helpText := `
Usage: sloppy show [OPTIONS] [PROJECT[/SERVICE[/APP]]]

   Outputs information for all projects or the given project, service or app

Options:

  -r, --raw       Prints raw json data

Examples:

  sloppy show
  sloppy show letschat
  sloppy show --raw letschat/frontend/nginx
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *ShowCommand) Run(args []string) int {
	var raw bool
	cmdFlags := newFlagSet("show", flag.ContinueOnError)
	cmdFlags.BoolVar(&raw, "r", false, "")
	cmdFlags.BoolVar(&raw, "raw", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy show --help'.")
		return 1
	}

	if code := c.UI.ErrorNoFlagAfterArg(cmdFlags.Args()); code == 1 {
		return code
	}

	var lastErr error
	var result interface{}
	if cmdFlags.NArg() == 0 {
		result, _, lastErr = c.Projects.List()
	} else {
		parts := strings.Split(strings.Trim(cmdFlags.Arg(0), "/"), "/")
		switch len(parts) {
		case 1:
			projects, _, err := c.Projects.Get(parts[0])
			if err != nil {
				lastErr = err
				break
			}
			result = projects.Services
		case 2:
			services, _, err := c.Services.Get(parts[0], parts[1])

			if err != nil {
				lastErr = err
				break
			}
			result = services.Apps
		case 3:
			result, _, lastErr = c.Apps.Get(parts[0], parts[1], parts[2])

		default:
			return c.UI.ErrorInvalidAppPath(cmdFlags.Arg(0))
		}
	}
	if lastErr != nil {
		c.UI.ErrorAPI(lastErr)
		return 1
	}

	if raw {
		return c.UI.Raw(result)
	}

	c.UI.Table("show", result)

	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *ShowCommand) Synopsis() string {
	return "Show settings of a project, a service or an application"
}
