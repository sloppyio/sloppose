package command

import (
	"strings"

	"github.com/sloppyio/cli/src/api"
	"github.com/sloppyio/cli/src/ui"
)

// DockerLogoutCommand is a Command implementation that deletes the docker
// credentials from sloppy.io
type DockerLogoutCommand struct {
	UI                  ui.UI
	RegistryCredentials api.RegistryCredentialsCheckDeleter
}

// Help should return long-form help text.
func (c *DockerLogoutCommand) Help() string {
	helpText := `
Usage: sloppy docker-logout

  Deletes docker credentials from sloppy.io
`

	return strings.TrimSpace(helpText)
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *DockerLogoutCommand) Run(args []string) int {
	if _, _, err := c.RegistryCredentials.Check(); err == api.ErrMissingAccessToken {
		c.UI.ErrorAPI(err)
		return 1
	} else if err != nil {
		c.UI.Error("You currently don't have access to private repositories.")
		return 1
	}

	if _, _, err := c.RegistryCredentials.Delete(); err != nil {
		c.UI.ErrorAPI(err)
		return 1
	}
	c.UI.Info("Removed access to your private repository.")
	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *DockerLogoutCommand) Synopsis() string {
	return "Removes docker credentials from sloppy.io"
}
