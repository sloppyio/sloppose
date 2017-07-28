package command

import (
	"flag"
	"strings"

	"sevenval.com/sloppose/pkg/converter"
)

type Convert struct{}

func (c *Convert) Help() string {
	text := `
	Usage: sloppose convert [options] [files]

	Defaults to docker-compose.yml if no files are given.

	Converts a docker-compose.yml to sloppyio.yml file.
	`
	return strings.TrimSpace(text)
}

func (c *Convert) Run(args []string) error {
	var output, projectName string
	flagSet := &flag.FlagSet{}
	flagSet.StringVar(&output, "o", "", "-o path/file.yml")
	flagSet.StringVar(&projectName, "projectname", "", "-projectname yourProjectName")
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	cf, err := converter.NewComposeFile(flagSet.Args(), projectName)
	if err != nil {
		return err
	}

	sf, err := converter.NewSloppyFile(cf)
	if err != nil {
		println(err)
		return err
	}

	linker := &converter.Linker{}
	linker.Resolve(cf, sf)

	if output == "" {
		output = strings.ToLower(sf.Project)
	}
	writer := &converter.YAMLWriter{}
	return writer.WriteFile(sf, output)
}
