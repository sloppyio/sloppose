package command

import "sevenval.com/sloppose/cli"

var Commands map[string]cli.CommandFactory

func init() {
	Commands = map[string]cli.CommandFactory{
		"convert": func() (cli.Command, error) {
			return &Convert{}, nil
		},
		"version": func() (cli.Command, error) {
			return &Version{}, nil
		},
	}
}
