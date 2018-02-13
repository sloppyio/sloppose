package main

import (
	"fmt"
	"os"

	"github.com/sloppyio/sloppose/command"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please specify a command...")
		usage()
		os.Exit(1)
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
			fmt.Printf("Error: %s\n\n", err.Error())
			fmt.Println(cmd.Help())
		}
	} else {
		usage()
	}
}

func fatal(err error) {
	fmt.Println("Error:", err.Error())
	os.Exit(1)
}

func usage() {
	var cmdStr string
	for cmd, cmdFn := range command.Commands {
		c, _ := cmdFn()
		cmdStr += fmt.Sprintf("  %s\t\t%s\n", cmd, c.Synopsis())
	}
	fmt.Printf(usageTemplate, cmdStr)
}

const usageTemplate = `Available commands:
%s
`
