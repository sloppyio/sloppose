package command

import (
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// RestartCommand is a Command implementation that is used to restart a
// specific application.
type RestartCommand struct {
	UI   ui.UI
	Apps api.AppsRestarter
}

// Help should return long-form help text.
func (c *RestartCommand) Help() string {
	helpText := `
Usage: sloppy restart PROJECT/SERVICE/APP

  Restarts the given app

Examples:

  sloppy restart letschat/frontend/apache
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *RestartCommand) Run(args []string) int {
	if len(args) < 1 {
		return c.UI.ErrorNotEnoughArgs("restart", "", 1)
	}

	parts := strings.Split(strings.Trim(args[0], "/"), "/")
	if len(parts) != 3 {
		return c.UI.ErrorInvalidAppPath(args[0])
	}

	status, _, err := c.Apps.Restart(parts[0], parts[1], parts[2])
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Info(status.Message)

	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *RestartCommand) Synopsis() string {
	return "Restart an app"
}
