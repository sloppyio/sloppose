package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestRollbackCommand_implements(t *testing.T) {
	c := &RollbackCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy rollback") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Rollback") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestRollbackCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &RollbackCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/node",
		"1234",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestRollbackCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &RollbackCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/node/node",
		"12345",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}

func TestRollbackCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &RollbackCommand{UI: mockUI}

	args := []string{}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 2 arguments")
}

func TestRollbackCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &RollbackCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache",
		"1234",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}
