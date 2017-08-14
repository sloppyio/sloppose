package test

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

type Helper struct {
	t         *testing.T
	initialWd string
	wd        string // WorkingDirectory
	td        string // TemporaryDirectory
}

func NewHelper(t *testing.T) *Helper {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return &Helper{
		t:         t,
		initialWd: cwd,
		wd:        cwd,
	}
}

func (h *Helper) GetTestFile(name string) io.ReadCloser {
	file, err := os.Open(filepath.Join(h.wd, "testdata", name))
	h.Must(err)
	return file
}

func (h *Helper) ChdirTemp() {
	tmpDir := os.TempDir()
	err := os.Chdir(tmpDir)
	h.Must(err)
	h.td = tmpDir
	h.wd = h.td
}

func (h *Helper) ChdirTest() {
	err := os.Chdir(h.initialWd)
	h.Must(err)
	h.td = ""
	h.wd = h.initialWd
	os.TempDir()
}

// Must ensures a test failure
func (h *Helper) Must(err error) {
	if err != nil {
		h.t.Fatalf("Err: %v", err)
	}
}
