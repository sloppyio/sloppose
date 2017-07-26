package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

func TestDockerLogoutCommand_implements(t *testing.T) {
	c := &DockerLogoutCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy docker-logout") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestDockerLogoutCommand(t *testing.T) {
	registryCredentials := &mockRegistryCredentialsEndpoint{}
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DockerLogoutCommand{UI: mockUI, RegistryCredentials: registryCredentials}

	testCodeAndOutput(t, mockUI, c.Run([]string{}), 0, "Removed access to your private repository.")
}

func TestDockerLogoutCommand_failed(t *testing.T) {
	registryCredentials := &mockRegistryCredentialsEndpoint{
		wantMessage: "No credentials found",
	}
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DockerLogoutCommand{UI: mockUI, RegistryCredentials: registryCredentials}

	testCodeAndOutput(t, mockUI, c.Run([]string{}), 1, "You currently don't have access to private repositories.")
}

func TestDockerLogoutCommand_notLoggedIn(t *testing.T) {
	registryCredentials := &mockRegistryCredentialsEndpoint{
		wantMessage: api.ErrMissingAccessToken.Error(),
	}
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DockerLogoutCommand{UI: mockUI, RegistryCredentials: registryCredentials}

	testCodeAndOutput(t, mockUI, c.Run([]string{}), 1, "not logged in")
}
