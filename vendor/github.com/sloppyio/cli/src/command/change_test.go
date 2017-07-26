package command

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

func TestChangeCommand_implements(t *testing.T) {
	c := &ChangeCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy change") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Change") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestChangeCommand_updateApp(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Apps: apps}

	args := []string{
		"-m", "1024",
		"-instances=2",
		"-image=node",
		"-d=test.sloppy.zone",
		"-env=test:12",
		"-env=foo:bar",
		"letschat/frontend/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")

	want := &api.App{
		Memory:    api.Int(1024),
		Instances: api.Int(2),
		Image:     api.String("node"),
		EnvVars: map[string]string{
			"test": "12",
			"foo":  "bar",
		},
		Domain: &api.Domain{
			URI: api.String("test.sloppy.zone"),
		},
	}

	if !reflect.DeepEqual(apps.input, want) {
		t.Errorf("Request = %+v, want %+v", apps.input, want)
	}
}

func TestChangeCommand_updateProject(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Projects: projects}

	extTests := []struct {
		ext string
	}{
		{ext: "json"},
		{ext: "yml"},
	}

	for _, tt := range extTests {
		args := []string{
			"-var=memory:1024",
			"-var=instances:1",
			"../../tests/files/letschat_variable." + tt.ext,
		}

		testCodeAndOutput(t, mockUI, c.Run(args), 0, "frontend")
	}
}

func TestChangeCommand_updateProjecIncorrectOrder(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Projects: projects}

	args := []string{
		"-var=memory:1024",
		"-var=instances:1",
		"../../tests/files/letschat_variable.json",
		"letschat",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "incorrect order of arguments")

}

func TestChangeCommand_updateProjectBackwardCompatibility(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Projects: projects}

	extTests := []struct {
		ext string
	}{
		{ext: "json"},
		{ext: "yml"},
	}

	for _, tt := range extTests {
		args := []string{
			"-var=memory:1024",
			"-var=instances:1",
			"letschat",
			"../../tests/files/letschat_variable." + tt.ext,
		}

		testCodeAndOutput(t, mockUI, c.Run(args), 0, "frontend")
	}
}

func TestChangeCommand_notEnoughArgsApp(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ChangeCommand{UI: mockUI}

	args := []string{}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestChangeCommand_notEnoughArgsAppNoFileNoAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ChangeCommand{UI: mockUI}

	args := []string{"noapppath"}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "application path or project file required")
}

func TestChangeCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ChangeCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/node/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "")
}

func TestChangeCommand_flagsAfterArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ChangeCommand{UI: mockUI}

	args := []string{
		"../../tests/files/testproject_variable.json",
		"--var=instances:1",
		"--var=memory:1",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "OPTIONS need to be set first")
}

func TestChangeCommand_missingVariableValues(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ChangeCommand{UI: mockUI}

	args := []string{
		"--var=instances:1",
		"../../tests/files/testproject_variable.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "missing variable 'memory'")
}

func TestChangeCommand_notFoundCreate(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Apps: apps, Projects: projects}

	args := []string{
		"--var=instances:1",
		"--var=memory:1",
		"../../tests/files/letschat_variable.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "frontend")
}

func TestChangeCommand_notFoundCreateBackwardCompatibility(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Apps: apps, Projects: projects}

	args := []string{
		"--var=instances:1",
		"--var=memory:1",
		"letschat",
		"../../tests/files/letschat_variable.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "frontend")
}

func TestChangeCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Apps: apps, Projects: projects}

	args := []string{
		"--i=1",
		"letschat/frontend/apache1",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}

func TestChangeCommand_missingOptions(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	projects := &mockProjectsEndpoint{}
	c := &ChangeCommand{UI: mockUI, Apps: apps, Projects: projects}

	args := []string{
		"letschat/frontend/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "missing options")
}
