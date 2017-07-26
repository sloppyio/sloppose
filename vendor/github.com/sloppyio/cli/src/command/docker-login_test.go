package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/sloppyio/cli/src/ui"
)

func TestDockerLoginCommand_implements(t *testing.T) {
	c := &DockerLoginCommand{}

	if !strings.Contains(c.Help(), "Usage: sloppy docker-login") {
		t.Errorf("Help = %s", c.Help())
	}

	if !strings.Contains(c.Synopsis(), "") {
		t.Errorf("Synopsis = %s", c.Synopsis())
	}
}

func TestDockerLoginCommand(t *testing.T) {
	inR, inW := io.Pipe()
	defer inR.Close()
	defer inW.Close()

	registryCredentials := &mockRegistryCredentialsEndpoint{}
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{InputReader: inR}}
	c := &DockerLoginCommand{UI: mockUI, RegistryCredentials: registryCredentials}

	// Create dummy file
	file := createTempFile(t, "docker", "success")
	defer os.Remove(file.Name())

	args := []string{file.Name()}

	go fmt.Fprintf(inW, "y\n")

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "to our service. You can now launch apps from your private repositories.")
}

func TestDockerLoginCommand_failed(t *testing.T) {
	inR, inW := io.Pipe()
	defer inR.Close()
	defer inW.Close()

	registryCredentials := &mockRegistryCredentialsEndpoint{}
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{InputReader: inR}}
	c := &DockerLoginCommand{UI: mockUI, RegistryCredentials: registryCredentials}

	// Create dummy file
	file := createTempFile(t, "docker", "failed")
	defer os.Remove(file.Name())

	args := []string{file.Name()}

	go fmt.Fprintf(inW, "y\n")

	testCodeAndOutput(t, mockUI, c.Run(args), 1, "Unable to upload docker credentials")
}

func TestDockerLoginCommand_abort(t *testing.T) {
	inR, inW := io.Pipe()
	defer inR.Close()
	defer inW.Close()

	mockUI := &ui.MockUI{MockUi: &cli.MockUi{InputReader: inR}}
	c := &DockerLoginCommand{UI: mockUI}

	// Create dummy file
	file := createTempFile(t, "docker", "abort")
	defer os.Remove(file.Name())

	args := []string{file.Name()}

	go fmt.Fprintf(inW, "n\n")

	testCodeAndOutput(t, mockUI, c.Run(args), 0, "")
}

func TestDockerLoginCommand_noDockerConfig(t *testing.T) {
	mockUI := &ui.MockUI{MockUi: &cli.MockUi{}}
	c := &DockerLoginCommand{UI: mockUI}

	args := []string{"noDockerConfig.json"}
	testCodeAndOutput(t, mockUI, c.Run(args), 1, "doesn't exist.")
}

// createTempFile creates a dummy file for testing purpose
func createTempFile(t *testing.T, name, content string) *os.File {
	file, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		t.Fatal("Couldn't create temp file!")
	}
	defer file.Close()

	file.Write([]byte(`{"auth":"` + content + `"}`))
	file.Sync()

	return file
}
