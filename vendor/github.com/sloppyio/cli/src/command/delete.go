package command

import (
	"flag"
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// DeleteCommand is a Command implementation that is used to delete a project,
// a service or an application.
type DeleteCommand struct {
	UI       ui.UI
	Projects api.ProjectsDeleter
	Services api.ServicesDeleter
	Apps     api.AppsDeleter
}

// Help should return long-form help text.
func (c *DeleteCommand) Help() string {
	helpText := `
Usage: sloppy delete [OPTIONS] PROJECT[/SERVICE[/APP]]

  Deletes the given project, service or application

Options:

  -f, --force=false   Force the deletion of a given project, service or an app

Examples:

  sloppy delete letschat
  sloppy delete -f letschat/frontend
  sloppy delete letschat/frontend/apache
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *DeleteCommand) Run(args []string) int {
	var force bool
	cmdFlags := newFlagSet("delete", flag.ContinueOnError)
	cmdFlags.BoolVar(&force, "f", false, "")
	cmdFlags.BoolVar(&force, "force", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy delete --help'.")
		return 1
	}

	if code := c.UI.ErrorNoFlagAfterArg(cmdFlags.Args()); code == 1 {
		return code
	}

	if cmdFlags.NArg() < 1 {
		return c.UI.ErrorNotEnoughArgs("delete", "", 1)
	}

	var status *api.StatusResponse
	var err error
	parts := strings.Split(strings.Trim(cmdFlags.Arg(0), "/"), "/")

	switch len(parts) {
	case 1:
		status, _, err = c.Projects.Delete(parts[0], force)
	case 2:
		status, _, err = c.Services.Delete(parts[0], parts[1], force)
	case 3:
		status, _, err = c.Apps.Delete(parts[0], parts[1], parts[2], force)
	default:
		return c.UI.ErrorInvalidAppPath(cmdFlags.Arg(0))
	}

	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Info(status.Message)
	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *DeleteCommand) Synopsis() string {
	return "Delete a project, a service or an application"
}
