package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// The git commit that was compiled. This will be filled in by the compiler.
var (
	GitCommit string
)

// Version number that is being run at the moment.
var Version string

// VersionPrerelease marks the version as pre-release. If this is ""
// (empty string) then it means that it is a final release. Otherwise, this
// is a pre-release such as "dev" (in development), "beta", "rc1", etc.
var VersionPrerelease string

// OS is the running cli's operating system.
var OS = strings.Title(strings.Replace(runtime.GOOS, "darwin", "macintosh", -1))

func userAgent() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "sloppy-cli/%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&buf, ".%s", VersionPrerelease)
	}
	fmt.Fprintf(&buf, " (%s) go/%s", OS, runtime.Version()[2:])

	return buf.String()
}
