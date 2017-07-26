package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/command"
	"github.com/sloppyio/cli/src/ui"
)

// Commands is the mapping of all the available sloppy commands.
var Commands map[string]cli.CommandFactory

func init() {
	// Add default UI
	defaultUI := ui.NewUI()

	Commands = map[string]cli.CommandFactory{
		"change": func() (cli.Command, error) {
			return &command.ChangeCommand{
				Projects: client.Projects,
				Apps:     client.Apps,
				UI:       defaultUI,
			}, nil
		},
		"delete": func() (cli.Command, error) {
			return &command.DeleteCommand{
				Projects: client.Projects,
				Services: client.Services,
				Apps:     client.Apps,
				UI:       defaultUI,
			}, nil
		},
		"docker-login": func() (cli.Command, error) {
			return &command.DockerLoginCommand{
				RegistryCredentials: client.RegistryCredentials,
				UI:                  defaultUI,
			}, nil
		},
		"docker-logout": func() (cli.Command, error) {
			return &command.DockerLogoutCommand{
				RegistryCredentials: client.RegistryCredentials,
				UI:                  defaultUI,
			}, nil
		},
		"logs": func() (cli.Command, error) {
			return &command.LogsCommand{
				Projects: client.Projects,
				Services: client.Services,
				Apps:     client.Apps,
				UI:       defaultUI,
			}, nil
		},
		"restart": func() (cli.Command, error) {
			return &command.RestartCommand{
				Apps: client.Apps,
				UI:   defaultUI,
			}, nil
		},
		"rollback": func() (cli.Command, error) {
			return &command.RollbackCommand{
				Apps: client.Apps,
				UI:   defaultUI,
			}, nil
		},
		"scale": func() (cli.Command, error) {
			return &command.ScaleCommand{
				Apps: client.Apps,
				UI:   defaultUI,
			}, nil
		},
		"show": func() (cli.Command, error) {
			return &command.ShowCommand{
				Projects: client.Projects,
				Services: client.Services,
				Apps:     client.Apps,
				UI:       defaultUI,
			}, nil
		},
		"start": func() (cli.Command, error) {
			return &command.StartCommand{
				Projects: client.Projects,
				UI:       defaultUI,
			}, nil
		},
		"stats": func() (cli.Command, error) {
			return &command.StatsCommand{
				Projects: client.Projects,
				Apps:     client.Apps,
				UI:       defaultUI,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				CheckVersion:      checkVersion,
				UI:                defaultUI,
			}, nil
		},
	}
}

// BasicHelpFunc generates some basic help output.
// Copied from cli, just change help flag order.
func BasicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"usage: %s [--version] <command> [<args>] [--help]\n\n",
			app))
		buf.WriteString("Available commands are:\n")

		// Get the list of keys so we can sort them, and also get the maximum
		// key length so they can be aligned properly.
		keys := make([]string, 0, len(commands))
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}

			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc, ok := commands[key]
			if !ok {
				// This should never happen since we JUST built the list of
				// keys.
				panic("command not found: " + key)
			}

			command, err := commandFunc()
			// Don't advertise every command globally
			if command.Synopsis() == "" {
				continue
			}

			if err != nil {
				log.Printf("[ERR] cli: Command '%s' failed to load: %s",
					key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}
