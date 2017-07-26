package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestRestartCommand_implements(t *testing.T) {
	c := &RestartCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy restart") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Restart") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestRestartCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &RestartCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/node",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 0, "Restarting app.")
}

func TestRestartCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &RestartCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/node/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}

func TestRestartCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &RestartCommand{UI: mockUI}

	args := []string{}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestRestartCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &RestartCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}
