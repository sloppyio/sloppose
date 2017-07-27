package main

import (
	"fmt"
	"os"

	"sevenval.com/sloppose/command"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fatal(fmt.Errorf("Please specify a command..."))
	}

	if args[0] == "-v" || args[0] == "--version" {
		args = []string{"version"}
	}

	// TODO move command handling to cli struct
	if _, ok := command.Commands[args[0]]; ok {
		cmd, err := command.Commands[args[0]]()
		if err != nil {
			fatal(err)
		}
		err = cmd.Run(args[1:])
		if err != nil {
			fmt.Printf(usageTemplate, cmd.Help())
		}
	} else {
		// TODO print possible commands
		fmt.Println("Please specify a command...")
		fmt.Printf(usageTemplate, "TODO")
	}

}

func fatal(err error) {
	fmt.Println("Error:", err.Error())
	os.Exit(1)
}

const usageTemplate = `Command usage:
	%s
`
