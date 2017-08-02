package command

import "fmt"

var VersionName = "v0.1.0"
var BuildName = "dev"

type Version struct{}

func (v *Version) Help() string {
	return v.Synopsis()
}

func (v *Version) Synopsis() string {
	return "Prints the current version and build hash."
}

func (v *Version) Run(args []string) error {
	fmt.Printf("Version: %s, Build: %s\n", VersionName, BuildName)
	return nil
}
