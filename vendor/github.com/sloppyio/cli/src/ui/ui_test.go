package ui

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mitchellh/cli"

	"github.com/sloppyio/cli/src/api"
)

func TestNewUI_implements(t *testing.T) {
	NewUI()
}

func TestErrorAPI(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}

	err := &api.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusConflict,
			Status:     http.StatusText(http.StatusConflict),
			Request:    &http.Request{},
		},
		StatusResponse: api.StatusResponse{
			Status:  "error",
			Message: "App is locked by one or more deployments.",
		},
		Reason: "Please wait for them to be finished before you trigger another one.",
	}

	mockUI.ErrorAPI(err)
	errOut := mockUI.ErrorWriter.String()
	want := err.Message + " " + err.Reason + "\n"
	if errOut != want {
		t.Errorf("ErrorAPI(%v) = %s, want %s", err, errOut, want)
	}
}

func TestErrorMissingAccessToken(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}

	mockUI.ErrorAPI(api.ErrMissingAccessToken)
	errOut := mockUI.ErrorWriter.String()
	if !strings.Contains(errOut, "not logged in") {
		t.Errorf("ErrorAPI(missingToken) = %s, want %s", errOut, "error: not logged in")
	}
}

func TestErrorNotEnoughArgs(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}

	if code := mockUI.ErrorNotEnoughArgs("update", "", 1); code != 1 {
		t.Errorf("ExitCode = %d, want %d", code, 1)
		t.Errorf("Output = %s", mockUI.OutputWriter.String())
	}
	err := mockUI.ErrorWriter
	if !strings.Contains(err.String(),
		"'update' requires a minimum of 1 argument") {
		t.Errorf("Error = %s", err.String())
	}
	mockUI.ErrorWriter.Reset()

	mockUI.ErrorNotEnoughArgs("scale", "", 2)
	if !strings.Contains(err.String(),
		"'scale' requires a minimum of 2 arguments") {
		t.Errorf("Error = %s", err.String())
	}
}

func TestErrorInvalidAppPath(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}

	if code := mockUI.ErrorInvalidAppPath("abc/def/ghe/aa"); code != 1 {
		t.Errorf("ExitCode = %d, want %d", code, 1)
		t.Errorf("Output = %s", mockUI.OutputWriter.String())
	}
	err := mockUI.ErrorWriter
	if !strings.Contains(err.String(),
		"invalid application path 'abc/def/ghe/aa'") {
		t.Errorf("Error = %s", err.String())
	}
}

func TestRaw(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}
	var raw = struct {
		Test string `json:"test"`
	}{
		Test: "foo",
	}
	if code := mockUI.Raw(raw); code != 0 {
		t.Errorf("ExitCode = %d, want %d", code, 0)
		t.Errorf("Error = %s", mockUI.ErrorWriter.String())
	}

	want := `{
    "test": "foo"
}` + "\n"

	out := mockUI.OutputWriter.String()
	if out != want {
		t.Errorf("Output = %s, want %s", out, want)
	}
}

func TestRaw_invalidStruct(t *testing.T) {
	mockUI := &MockUI{&cli.MockUi{}}
	var raw = map[interface{}]interface{}{
		"test": "string",
	}

	if code := mockUI.Raw(raw); code != 1 {
		t.Errorf("ExitCode = %d, want %d", code, 1)
		t.Errorf("Output = %s", mockUI.OutputWriter.String())
	}
}
