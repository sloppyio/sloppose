package command

import (
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// RollbackCommand is a Command implementation that is used to roll back an
// application to a specific version.
type RollbackCommand struct {
	UI   ui.UI
	Apps api.AppsRollbacker
}

// Help should return long-form help text.
func (c *RollbackCommand) Help() string {
	helpText := `
Usage: sloppy rollback PROJECT/SERVICE/APP VERSION

  Allows you to roll back an application to a specific version

Examples:

  sloppy rollback letschat/frontend/apache 2015-11-18T16:56:30.495Z
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *RollbackCommand) Run(args []string) int {
	if len(args) < 2 {
		return c.UI.ErrorNotEnoughArgs("rollback", "", 2)
	}

	parts := strings.Split(strings.Trim(args[0], "/"), "/")
	if len(parts) != 3 {
		return c.UI.ErrorInvalidAppPath(args[0])
	}

	app, _, err := c.Apps.Rollback(parts[0], parts[1], parts[2], args[1])
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Table("show", app)
	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *RollbackCommand) Synopsis() string {
	return "Rollback an application"
}
