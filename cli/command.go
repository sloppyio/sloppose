package cli

type Command interface {
	Help() string
	Run(args []string) error
}
