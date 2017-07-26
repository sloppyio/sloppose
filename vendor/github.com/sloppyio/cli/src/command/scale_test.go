package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestScaleCommand_implements(t *testing.T) {
	c := &ScaleCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy scale") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Scale") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestScaleCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &ScaleCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/node",
		"1",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestScaleCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ScaleCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/node/node",
		"2",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}

func TestScaleCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ScaleCommand{UI: mockUI}

	args := []string{}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 2 arguments")
}

func TestScaleCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &ScaleCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache",
		"1",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}

func TestScaleCommand_invalidArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	apps := &mockAppsEndpoint{}
	c := &ScaleCommand{UI: mockUI, Apps: apps}

	args := []string{
		"letschat/frontend/apache",
		"abc",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid instance number")
}
