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

	sloppyService, err := converter.NewSloppyFile(cf)
	if err != nil {
		println(err)
		return err
	}

	writer := &converter.YAMLWriter{}
	return writer.WriteFile(sloppyService, "out.yml")
}
