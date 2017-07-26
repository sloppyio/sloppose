package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestDeleteCommand_implements(t *testing.T) {
	c := &DeleteCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy delete") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Delete") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestDeleteCommand_deleteProject(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &DeleteCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "Project letschat successfully deleted.")
}

func TestDeleteCommand_deleteService(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &DeleteCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "Service frontend successfully deleted.")
}

func TestDeleteCommand_deleteApp(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &DeleteCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "App node successfully deleted.")
}

func TestDeleteCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &DeleteCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend/node1",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}

func TestDeleteCommand_flagsAfterArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &DeleteCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend/node",
		"-f",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "OPTIONS need to be set first")
}

func TestDeleteCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DeleteCommand{UI: mockUI}

	args := []string{}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestDeleteCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DeleteCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/apache/apache",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}
