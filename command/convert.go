package command

import (
	"flag"
	"strings"

	"github.com/sloppyio/sloppose/pkg/converter"
)

type Convert struct{}

func (c *Convert) Help() string {
	text := `
Usage: sloppose convert [options] [files]

Options:
  -o              output path, defaults to working directory
  -projectname    sets the projectname, defaults to working directory

Defaults to docker-compose.yml if no files are given.
Converts a docker-compose.yml to a sloppy.io compatible yml format.
`
	return strings.TrimSpace(text)
}

func (c *Convert) Synopsis() string {
	return "Converts a docker-compose.yml to a sloppy.io compatible yml format."
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

	reader := &converter.ComposeReader{}
	buf, err := reader.ReadAll(flagSet.Args())
	if err != nil {
		return err
	}

	cf, err := converter.NewComposeFile(buf, projectName)
	if err != nil {
		return err
	}

	sf, err := converter.NewSloppyFile(cf)
	if err != nil {
		println(err)
		return err
	}

	linker := &converter.Linker{}
	err = linker.Resolve(cf, sf)
	if err != nil {
		return err
	}

	if output == "" {
		output = strings.ToLower(sf.Project)
	}
	writer := &converter.YAMLWriter{}
	return writer.WriteFile(sf, output)
}
