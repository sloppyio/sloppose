// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"testing"
	"time"
)

func TestStringify(t *testing.T) {
	var nilPointer *string

	var tests = []struct {
		in  interface{}
		out string
	}{
		// basic types
		{"foo", `"foo"`},
		{123, `123`},
		{1.5, `1.5`},
		{false, `false`},
		{
			[]string{"a", "b"},
			`["a" "b"]`,
		},
		{
			struct {
				A []string
			}{nil},
			// nil slice is skipped
			`{}`,
		},
		{
			struct {
				A string
			}{"foo"},
			// structs not of a named type get no prefix
			`{A:"foo"}`,
		},

		// pointers
		{nilPointer, `<nil>`},
		{String("foo"), `"foo"`},
		{Int(123), `123`},
		{Bool(false), `false`},
		{
			[]*string{String("a"), String("b")},
			`["a" "b"]`,
		},

		// actual GitHub structs
		{
			Timestamp{time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC)},
			`api.Timestamp{2006-01-02 15:04:05 +0000 UTC}`,
		},
		{
			&Timestamp{time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC)},
			`api.Timestamp{2006-01-02 15:04:05 +0000 UTC}`,
		},
		{
			App{ID: String("apache"), Instances: Int(5)},
			`api.App{ID:"apache", Instances:5, EnvVars:map[]}`,
		},
	}

	for i, tt := range tests {
		s := Stringify(tt.in)
		if s != tt.out {
			t.Errorf("%s", s)
			t.Errorf("%d. Stringify(%q) => %q, want %q", i, tt.in, s, tt.out)
		}
	}
}
