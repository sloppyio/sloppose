package command

import "fmt"

var VersionName = "v0.0.1"
var BuildName = "dev"

type Version struct{}

func (v *Version) Help() string {
	return ""
}

func (v *Version) Run(args []string) error {
	fmt.Printf("Version: %s, Build: %s\n", VersionName, BuildName)
	return nil
}
