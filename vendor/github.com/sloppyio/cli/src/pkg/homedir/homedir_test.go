package homedir

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	home := Get()
	if home == "" {
		t.Fatal("Expected non-empty home directory")
	}
}

func TestGet_emptyEnv(t *testing.T) {
	os.Setenv("HOME", "")
	home := Get()
	if home == "" {
		t.Fatal("Expected non-empty home directory")
	}
}
