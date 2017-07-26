package command

import (
	"reflect"
	"testing"
)

func TestStringMap(t *testing.T) {
	var m stringMap

	if err := m.Set("memory:128"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := m.Set("limit:,volume:1,instances:1"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	var want = stringMap{
		"memory":    "128",
		"instances": "1",
		"volume":    "1",
	}

	if !reflect.DeepEqual(want, m) {
		t.Errorf("stringMap Set: %v, want %v", m, want)
	}

}
