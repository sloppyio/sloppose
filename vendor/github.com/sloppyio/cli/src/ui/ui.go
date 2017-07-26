package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sloppyio/cli/src/api"

	"github.com/mitchellh/cli"
)

// UI is an interface for interacting with the terminal, or "interface" of
// a CLI. Based on cli.Ui
type UI interface {
	cli.Ui

	// ErrorAPI is used for api related error messages that might appear on
	// stderr.
	ErrorAPI(error)
	ErrorNotEnoughArgs(string, string, int) int
	ErrorInvalidAppPath(string) int
	ErrorNoFlagAfterArg([]string) int
	Table(string, interface{})
	Raw(interface{}) int
}

// DefaultUI is an implementation of UI.
type DefaultUI struct {
	cli.Ui
}

// NewUI returns a defaultUI for interacting with the terminal.
// DefaultUI is prefixing errors with "error: " and colorizing
// errors red and info green.
func NewUI() UI {
	var ui cli.Ui
	ui = &cli.BasicUi{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	// Disable colors windows
	if runtime.GOOS != "windows" {
		ui = &cli.ColoredUi{
			ErrorColor:  cli.UiColorRed,
			InfoColor:   cli.UiColorGreen,
			OutputColor: cli.UiColorNone,
			WarnColor:   cli.UiColorNone,
			Ui:          ui,
		}
	}

	return &DefaultUI{
		&cli.PrefixedUi{
			ErrorPrefix: "error: ",
			WarnPrefix:  "warning: ",
			Ui:          ui,
		},
	}
}

// ErrorAPI is used for api related error messages that might appear on
// stderr.
func (ui *DefaultUI) ErrorAPI(err error) {
	errStr := err.Error()
	if err == api.ErrMissingAccessToken {
		ui.Error("not logged in")
		return
	}

	if err, ok := err.(*api.ErrorResponse); ok {
		if err.Message == "" {
			errStr = err.Error()
		} else {
			errStr = fmt.Sprintf("%s %s", err.Message, err.Reason)
		}
	}

	ui.Error(errStr)
}

// ErrorNoFlagAfterArg is used for cli error related to wrong argument order
func (ui *DefaultUI) ErrorNoFlagAfterArg(s []string) int {
	for _, value := range s {
		if strings.HasPrefix(value, "-") {
			ui.Error("OPTIONS need to be set first.")
			return 1
		}
	}
	return 0
}

// ErrorNotEnoughArgs is used for cli error related to a minimum of argument.
func (ui *DefaultUI) ErrorNotEnoughArgs(command, help string, n int) int {
	plural := ""
	if n > 1 {
		plural = "s"
	}

	ui.Error(
		fmt.Sprintf("'%s' requires a minimum of %d argument%s.",
			command, n, plural),
	)

	if help == "" {
		help = "See 'sloppy " + command + " --help'."
	}

	ui.Output(help)
	return 1
}

// ErrorInvalidAppPath is used for cli error related to an invalid app path
func (ui *DefaultUI) ErrorInvalidAppPath(arg string) int {
	ui.Error(fmt.Sprintf("invalid application path '%s'. \n", arg))

	return 1
}

// Table prints tables depending on its given interface.
func (ui *DefaultUI) Table(kind string, v interface{}) {
	var buf bytes.Buffer
	switch kind {
	case "start":
		tableStart(&buf, v.([]*api.Service))
	case "show":
		tableShow(&buf, v)
	}

	ui.Output(buf.String())
}

// Raw prints the json representation of a given struct.
func (ui *DefaultUI) Raw(v interface{}) int {
	raw, err := json.MarshalIndent(v, "", "    ")

	if err != nil {
		ui.Error(fmt.Sprintf("Couldn't encode result: %v", err))
		return 1
	}

	ui.Output(string(raw))

	return 0
}

// MockUI is a mock UI that is used for tests and is exported publicly for
// use in external tests if needed as well.
type MockUI struct {
	*cli.MockUi
}

// ErrorAPI is used for api related error messages that might appear on
// stderr.
func (ui *MockUI) ErrorAPI(err error) {
	realUI := &DefaultUI{ui}
	realUI.ErrorAPI(err)
}

// ErrorNotEnoughArgs is used for cli error related to a minimum of argument.
func (ui *MockUI) ErrorNotEnoughArgs(command, help string, n int) int {
	realUI := &DefaultUI{ui}
	return realUI.ErrorNotEnoughArgs(command, help, n)
}

// Table is used for print a table
func (ui *MockUI) Table(kind string, v interface{}) {
	realUI := &DefaultUI{ui}
	realUI.Table(kind, v)
}

// Raw is used for print json
func (ui *MockUI) Raw(v interface{}) int {
	realUI := &DefaultUI{ui}
	return realUI.Raw(v)
}

// ErrorInvalidAppPath used for print error
func (ui *MockUI) ErrorInvalidAppPath(arg string) int {
	realUI := &DefaultUI{ui}
	return realUI.ErrorInvalidAppPath(arg)
}

// ErrorNoFlagAfterArg used for print error
func (ui *MockUI) ErrorNoFlagAfterArg(s []string) int {
	realUI := &DefaultUI{ui}
	return realUI.ErrorNoFlagAfterArg(s)
}
