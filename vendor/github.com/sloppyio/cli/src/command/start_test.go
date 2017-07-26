package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestStartCommand_implements(t *testing.T) {
	c := &StartCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy start") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "Start") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestStartCommand(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StartCommand{UI: mockUI, Projects: projects}

	extTests := []struct {
		ext string
	}{
		{ext: "json"},
		{ext: "yml"},
	}

	for _, tt := range extTests {
		args := []string{
			"-v", "memory:128",
			"-var=instances:1",
			"../../tests/files/testproject_variable." + tt.ext,
		}
		testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
	}
}

func TestStartCommand_oneVarsFlag(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StartCommand{UI: mockUI, Projects: projects}

	args := []string{
		"--var=memory:128,instances:1",
		"../../tests/files/testproject_variable.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestStartCommand_notEnoughArgs(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	args := []string{}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "minimum of 1 argument")
}

func TestStartCommand_notExistFile(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	args := []string{
		"nofile.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "file 'nofile.json' not found.")
}

func TestStartCommand_notSupportedFileExtension(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	file, err := os.Create(filepath.Join(os.TempDir(), "a.notsupported"))
	if err != nil {
		t.Fatalf("could not create file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	args := []string{
		file.Name(),
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "file extension not supported")
}

func TestStartCommand_missingVariableValues(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	args := []string{
		"--var=instances:1",
		"../../tests/files/testproject_variable.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "missing variable 'memory'")
}

func TestStartCommand_flagsAfterArgument(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	args := []string{
		"../../tests/files/testproject_variable.json",
		"--var=instances:1",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "OPTIONS need to be set first")
}

func TestStartCommand_invalidInputParameter(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StartCommand{UI: mockUI, Projects: projects}

	args := []string{
		"../../tests/files/testproject_invalid.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "missing the required")
}

func TestStartCommand_invalidJSON(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	projects := &mockProjectsEndpoint{}
	c := &StartCommand{UI: mockUI, Projects: projects}

	args := []string{
		"../../tests/files/testproject_invalidjson.json",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "syntax error around line 13:21")
}

func TestStartCommand_incorrectFlags(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &StartCommand{UI: mockUI}

	args := []string{
		"-unknown",
	}

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "")
}
