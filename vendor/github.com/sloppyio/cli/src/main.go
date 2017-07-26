package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/api"
)

const (
	envHost  = "SLOPPY_APIHOST"
	envToken = "SLOPPY_APITOKEN"
)

// client is used in each command to handle api requests.
var client *api.Client

func main() {
	stackTrace := false // stackTrace holds the state whether a stack trace is displayed
	defer func() {
		if err := recover(); err != nil {
			printError(os.Stderr, "Error executing CLI: %s", err)
			if stackTrace {
				debug.PrintStack()
			}
			os.Exit(1)
		}
	}()

	// Shortcut --version, -v to show version command.
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
		if arg == "--help" {
			args = append([]string{"--help"}, args...)
		}
		if arg == "--debug" {
			stackTrace = true
			args = append(args[:i], args[i+1:]...)
		}
	}

	// Update mechanism
	update := make(chan struct{}, 1)
	if len(args) > 0 && args[0] == "version" {
		update <- struct{}{}
	} else {
		go func() {
			if ok, output := checkVersion(); ok {
				fmt.Fprint(os.Stderr, output)
			}
			update <- struct{}{}
		}()
	}

	client = api.NewClient()
	client.UserAgent = userAgent()
	client.SetAccessToken(os.Getenv(envToken))
	host := os.Getenv(envHost)
	if err := client.SetBaseURL(host); err != nil {
		printError(os.Stderr, "error: parsing SLOPPY_APIHOST: %s\n", err.Error())
		os.Exit(1)
	}

	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
		HelpFunc: BasicHelpFunc("sloppy"),
	}

	exitCode, err := cli.Run()
	if err != nil {
		printError(os.Stderr, "Error executing CLI: %s\n", err.Error())
		exitCode = 1
	}

	<-update // wait for update goroutine
	os.Exit(exitCode)
}

func printError(w io.Writer, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if runtime.GOOS == "windows" {
		fmt.Fprint(w, message)
	} else {
		fmt.Fprintf(w, "\033[0;31m%s\033[0m\n", message)
	}
}
