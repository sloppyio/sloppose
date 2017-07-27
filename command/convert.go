package command

import (
	"sevenval.com/sloppose/pkg/converter"
)

type Convert struct{}

func (c *Convert) Help() string {
	return ""
}

func (c *Convert) Run(args []string) error {
	cf, err := converter.NewComposeFile(args)
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

	writer := &converter.YAMLWriter{}
	return writer.WriteFile(sf, "out.yml")
}
