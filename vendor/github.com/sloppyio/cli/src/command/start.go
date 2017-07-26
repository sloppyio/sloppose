package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// StartCommand is a Command implementation that is used to create a project
// along with all its services and applications.
type StartCommand struct {
	UI       ui.UI
	Projects api.ProjectsCreater
}

// Help should return long-form help text.
func (c *StartCommand) Help() string {
	helpText := `
Usage: sloppy start [OPTIONS] FILENAME

  Start a new project on the sloppy service

Options:
  -v, --var=[]     values to set for placeholders

Examples:

  sloppy start --var=domain:mydomain.sloppy.zone --var=memory:128 myproject.json
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *StartCommand) Run(args []string) int {
	var vars stringMap
	cmdFlags := newFlagSet("start", flag.ContinueOnError)
	cmdFlags.Var(&vars, "v", "")
	cmdFlags.Var(&vars, "var", "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy start --help'.")
		return 1
	}

	if code := c.UI.ErrorNoFlagAfterArg(cmdFlags.Args()); code == 1 {
		return code
	}

	if cmdFlags.NArg() < 1 {
		return c.UI.ErrorNotEnoughArgs("start", "", 1)
	}

	file, err := os.Open(cmdFlags.Arg(0))
	if err != nil {
		if os.IsNotExist(err) {
			c.UI.Error(fmt.Sprintf("file '%s' not found.", cmdFlags.Arg(0)))
		} else if os.IsPermission(err) {
			c.UI.Error(fmt.Sprintf("no read permission '%s'.", cmdFlags.Arg(0)))
		} else {
			c.UI.Error(err.Error())
		}
		return 1
	}
	defer file.Close()

	decoder := newDecoder(file, vars)
	var input = new(api.Project)

	ext := filepath.Ext(file.Name())
	switch ext {
	case ".json":
		if err := decoder.DecodeJSON(input); err != nil {
			c.UI.Error(fmt.Sprintf("failed to parse JSON file %s\n%s", file.Name(), err.Error()))
			return 1
		}
	case ".yaml", ".yml":
		if err := decoder.DecodeYAML(input); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
	default:
		c.UI.Error("file extension not supported, must be json or yaml.")
		return 1
	}

	project, _, err := c.Projects.Create(input)
	if err != nil {

		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Table("show", project.Services)
	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *StartCommand) Synopsis() string {
	return "Start a new project on the sloppy service"
}
