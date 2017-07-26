package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestStatsCommand_implements(t *testing.T) {
	c := &StatsCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy stats") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Display metrics") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestStatsCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &StatsCommand{UI: mockUI, Projects: projects, Apps: apps}

	args := []string{
		"letschat",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "")
	out := mockUI.OutputWriter.String()
	if !strings.Contains(out, "No apps running") {
		t.Errorf("Output = %s", out)
	}
}

func TestStatsCommand_withAllFlag(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &StatsCommand{UI: mockUI, Projects: projects, Apps: apps}

	args := []string{
		"--all",
		"letschat",
	}
	testCodeAndOutput(t, mockUI, c.Run(args), 0,
		"frontend/node-59f7ed 	 128 MiB / 1024 MiB 	 12.5% 	 5.5 MiB / 146 MiB 	 97.8 B / 106 KiB 	 12.9%")
}

func TestStatsCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StatsCommand{UI: mockUI}

	args := []string{}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestStatsCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StatsCommand{UI: mockUI, Projects: projects}

	args := []string{
		"abc",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}

func TestStatsCommand_invalidProjectPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StatsCommand{UI: mockUI, Projects: projects}

	args := []string{
		"abc/def",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid project path")
}
