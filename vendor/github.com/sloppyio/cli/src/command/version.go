package command

import (
	"bytes"
	"fmt"

	"github.com/sloppyio/cli/src/ui"
)

// VersionCommand is a Command implementation that prints the version.
type VersionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	CheckVersion      func() (bool, string)
	UI                ui.UI
}

// Help should return long-form help text.
func (c *VersionCommand) Help() string {
	return ""
}

// Run should run the actual command with the given CLI instance and
// command-line args.
func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "Sloppy %s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.UI.Output(versionString.String())
	if ok, output := c.CheckVersion(); ok {
		c.UI.Output(output)
	}

	return 0
}

// Synopsis should return a one-line, short synopsis of the command.
func (c *VersionCommand) Synopsis() string {
	return "Prints the sloppy version"
}
