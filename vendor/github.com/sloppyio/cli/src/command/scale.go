package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// ScaleCommand is a Command implementation that is used to scale the number
// of instances a specific application is running on.
type ScaleCommand struct {
	UI   ui.UI
	Apps api.AppsScaler
}

// Help should return long-form help text.
func (c *ScaleCommand) Help() string {
	helpText := `
Usage: sloppy scale PROJECT/SERVICE/APP INSTANCES

  Allows you to scale the number of instances a specific application is running
  on

Examples:

  sloppy scale letschat/frontend/apache 3
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *ScaleCommand) Run(args []string) int {
	if len(args) < 2 {
		return c.UI.ErrorNotEnoughArgs("scale", "", 2)
	}

	parts := strings.Split(strings.Trim(args[0], "/"), "/")
	if len(parts) != 3 {
		return c.UI.ErrorInvalidAppPath(args[0])
	}

	instances, err := strconv.Atoi(args[1])
	if err != nil {
		c.UI.Error(fmt.Sprintf("invalid instance number '%s'. \n", args[1]))
		c.UI.Output(c.Help())
		return 1
	}

	app, _, err := c.Apps.Scale(parts[0], parts[1], parts[2], instances)
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Table("show", app)
	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *ScaleCommand) Synopsis() string {
	return "Scale the number of instances in an application"
}
