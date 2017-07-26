package command

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

func TestShowCommand_implements(t *testing.T) {
	c := &ShowCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy show") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Show") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestShowCommand_printProjects(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	testCodeAndOutput(t, mockUI, c.Run(nil), 0, "")
}

func TestShowCommand_printServices(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")

	// TODO:
}

func TestShowCommand_printApps(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestShowCommand_printApp(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"letschat/frontend/node",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestShowCommand_printRaw(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"-r",
		"letschat",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")

	var got []*api.Service
	if err := json.NewDecoder(mockUI.OutputWriter).Decode(&got); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, mockProject.Services) {
		t.Errorf("Output = %s", mockUI.OutputWriter.String())
	}
}

func TestShowCommand_invalidAppPath(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ShowCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/apache/apache",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "invalid app")
}

func TestShowCommand_flagsAfterArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &ShowCommand{UI: mockUI}

	args := []string{
		"letschat/frontend/apache/apache",
		"--raw",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "OPTIONS need to be set first")
}

func TestShowCommand_notFound(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	services := &mockServicesEndpoint{}
	apps := &mockAppsEndpoint{}
	c := &ShowCommand{UI: mockUI, Projects: projects, Services: services, Apps: apps}

	args := []string{
		"abc/def",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
	// Reset Buffer
	mockUI.ErrorWriter.Reset()
	mockUI.OutputWriter.Reset()

	args = []string{
		"abc",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "not be found")
}
