package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateFile(t *testing.T) {
	updateFilename = updateFilename + "_testing"
	updateFilepath := filepath.Join(os.TempDir(), updateFilename)
	defer os.Remove(updateFilepath)

	if ok, err := updateFile(true); ok {
		t.Errorf("updateFile(true) = %t", ok)
	} else if err != nil {
		t.Errorf("updateFile error %v", err)
	}

	data, err := ioutil.ReadFile(updateFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = time.Parse(updateFormat, string(data)); err != nil {
		t.Errorf("File(%v): %v", string(data), err)
	}

	err = ioutil.WriteFile(updateFilepath, []byte(time.Now().Add(-25*time.Hour).Format(updateFormat)), 0600)
	if err != nil {
		t.Fatal(err)
	}

	if ok, _ := updateFile(true); ok {
		t.Errorf("updateFile(true): %t", ok)
	}

	// New version found
	updateFile(false)
	if data, _ := ioutil.ReadFile(updateFilepath); string(data) != "" {
		t.Errorf("File is not empty: %s", string(data))
	}
}

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{
			a:    "1",
			b:    "1.0",
			want: 0,
		},
		{
			a:    "1.1",
			b:    "1.0",
			want: 1,
		},
		{
			a:    "1.1",
			b:    "1.2.1",
			want: -1,
		},
		{
			a:    "a.b.1",
			b:    "0.0.1",
			want: -1,
		},
		{
			a:    "1.2.0-rc.1",
			b:    "1.2.1-rc.1",
			want: -1,
		},
		{
			a:    "1.2.1-rc.1",
			b:    "1.2.1",
			want: -1,
		},
	}

	for i, tt := range tests {
		if got := compareVersion(tt.a, tt.b); got != tt.want {
			t.Errorf("%d) compareVersion(%s,%s) = %d, want %d", i, tt.a, tt.b, got, tt.want)
		}
	}
}
