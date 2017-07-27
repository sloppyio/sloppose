package cli

type CommandFactory func() (Command, error)
