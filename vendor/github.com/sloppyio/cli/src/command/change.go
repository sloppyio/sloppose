package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// ChangeCommand is a Command implementation that changes the configuration
// of a specific application on the fly.
type ChangeCommand struct {
	UI       ui.UI
	Projects interface {
		api.ProjectsUpdater
		api.ProjectsGetter
		api.ProjectsCreater
	}
	Apps api.AppsUpdater
}

// Help should return long-form help text.
func (c *ChangeCommand) Help() string {
	helpText := `
Usage: sloppy change [OPTIONS] (PROJECT/SERVICE/APP | [PROJECT] FILENAME)

  Sets the new values for the given app.

Options:

  -m, --memory          the amount of memory the app should use
  -i, --instances       the number of instances the app should use
  -img, --image         the new image the app should use
  -d, --domain          the new domain name the app should use
  -e, --env=[]          set environment variables
  -v, --var=[]          values to set for placeholders
  -f, --force           set force flag

Examples:

  sloppy change -m 128 letschat/frontend/apache
  sloppy change -var=domain:abc.sloppy.zone letschat.json
`
	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *ChangeCommand) Run(args []string) int {
	if len(args) == 0 {
		return c.UI.ErrorNotEnoughArgs("change", "", 1)
	}
	lastArg := args[len(args)-1]

	validExt := func(arg string) bool {
		return strings.HasSuffix(arg, ".json") || strings.HasSuffix(arg, ".yml") || strings.HasSuffix(arg, ".yaml")
	}

	switch {
	case strings.Count(strings.Trim(lastArg, "/"), "/") == 2 && !validExt(lastArg):
		return c.updateApp(args)
	case validExt(lastArg):
		return c.updateProject(args)
	case strings.HasPrefix(lastArg, "-"):
		return c.UI.ErrorNoFlagAfterArg(args)
	case len(args) >= 2:
		if validExt(args[len(args)-2]) {
			c.UI.Error("incorrect order of arguments.")
			c.UI.Output("See 'sloppy change --help'.")
			return 1
		}
	}
	c.UI.Error("application path or project file required.")
	return 1
}

// Update an application
func (c *ChangeCommand) updateApp(args []string) int {
	var memory, instances int
	var image, domain string
	var env stringMap
	cmdFlags := newFlagSet("change", flag.ContinueOnError)
	cmdFlags.IntVar(&memory, "m", 0, "")
	cmdFlags.IntVar(&memory, "memory", 0, "")
	cmdFlags.IntVar(&instances, "i", -1, "")
	cmdFlags.IntVar(&instances, "instances", -1, "")
	cmdFlags.Var(&env, "e", "")
	cmdFlags.Var(&env, "env", "")
	cmdFlags.StringVar(&image, "img", "", "")
	cmdFlags.StringVar(&image, "image", "", "")
	cmdFlags.StringVar(&domain, "d", "", "")
	cmdFlags.StringVar(&domain, "domain", "", "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy change --help'.")
		return 1
	}

	if code := c.UI.ErrorNoFlagAfterArg(cmdFlags.Args()); code == 1 {
		return code
	}

	parts := strings.Split(strings.Trim(cmdFlags.Arg(0), "/"), "/")
	if len(parts) != 3 {
		return c.UI.ErrorInvalidAppPath(cmdFlags.Arg(0))
	}

	app := new(api.App)
	if memory != 0 {
		app.Memory = api.Int(memory)
	}
	if instances >= 0 {
		app.Instances = api.Int(instances)
	}
	if image != "" {
		app.Image = api.String(image)
	}
	if domain != "" {
		app.Domain = &api.Domain{URI: api.String(domain)}
	}

	if env != nil {
		app.EnvVars = env
	}

	if reflect.DeepEqual(app, new(api.App)) {
		c.UI.Error("missing options.")
		return 1
	}

	app, _, err := c.Apps.Update(parts[0], parts[1], parts[2], app)
	if err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}

	c.UI.Table("show", app)

	return 0
}

// Updates an entire project
func (c *ChangeCommand) updateProject(args []string) int {
	var vars stringMap
	var force bool
	cmdFlags := newFlagSet("change", flag.ContinueOnError)
	cmdFlags.Var(&vars, "var", "")
	cmdFlags.Var(&vars, "v", "")
	cmdFlags.BoolVar(&force, "f", false, "")
	cmdFlags.BoolVar(&force, "force", false, "")

	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		c.UI.Output("See 'sloppy change --help'.")
		return 1
	}

	usedChainedVars := false
	for _, value := range args {
		if strings.Contains(value, ",") {
			usedChainedVars = true
		}
	}
	if usedChainedVars {
		c.UI.Warn("var chained with comma are deprecated.")
	}

	if code := c.UI.ErrorNoFlagAfterArg(cmdFlags.Args()); code == 1 {
		return code
	}

	filename := cmdFlags.Arg(0)
	var projectName string
	if cmdFlags.NArg() == 2 {
		c.UI.Warn("set project name explicitly is deprecated.")
		projectName = cmdFlags.Arg(0)
		filename = cmdFlags.Arg(1)
	} else if cmdFlags.NArg() < 1 {
		return c.UI.ErrorNotEnoughArgs("change", "", 1)
	}

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.UI.Error(fmt.Sprintf("file '%s' not found.", filename))
		} else if os.IsPermission(err) {
			c.UI.Error(fmt.Sprintf("no read permission '%s'.", filename))
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
		c.UI.Error("file extension not supported, must be json or yaml")
		return 1
	}

	var project *api.Project
	if projectName == "" {
		projectName = *input.Name
	}
	if _, _, err := c.Projects.Get(projectName); err != nil {
		project, _, err = c.Projects.Create(input)
		if err != nil {
			c.UI.ErrorAPI(err)
			return 1
		}
	} else {
		project, _, err = c.Projects.Update(projectName, input, force)
		if err != nil {
			c.UI.ErrorAPI(err)
			return 1
		}
	}

	c.UI.Table("show", project.Services)

	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *ChangeCommand) Synopsis() string {
	return "Change the configuration of an application on the fly"
}
