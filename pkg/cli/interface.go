package cli

import (
	"errors"
	"flag"
)

type Command struct {
	Run         func(args []string) error
	Flags       func() *flag.FlagSet
	Synopsis    string
	Subcommands Subcommands
}

var (
	ErrNotImplemented = errors.New("not implemented")
)

type CommandFunc func() Command

type Subcommands map[string]CommandFunc

func (c Command) Execute(args []string) error {
	return runCLI(func() Command { return c }, args)
}
