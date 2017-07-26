package command

import (
	"strings"
	"testing"

	"github.com/sloppyio/cli/src/ui"

	"github.com/mitchellh/cli"
)

func TestLogsCommand_implements(t *testing.T) {
	c := &LogsCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy logs") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Fetch the logs") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestLogsCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &LogsCommand{UI: mockUI, Projects: projects}

	args := []string{
		"letschat",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestLogsCommand_getLogsServices(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	services := &mockServicesEndpoint{}
	c := &LogsCommand{UI: mockUI, Services: services}

	args := []string{
		"letschat/frontend",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestLogsCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &LogsCommand{UI: mockUI}

	args := []string{}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestLogsCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &LogsCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}

func TestLogsCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &LogsCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}

func TestLogsCommand_flagsAfterArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &LogsCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache/node",
		"-n=5",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "OPTIONS need to be set first")
}
