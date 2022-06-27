package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/actions-go-build/internal/commands"
)

func main() {
	c, err := getCommand(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if err := c(); err != nil {
		log.Fatal(err)
	}
}

var errUsage = fmt.Errorf("usage: %s <command>", os.Args[0])

type command func(args []string) error
type bakedCommand func() error

func getCommand(osArgs []string) (bakedCommand, error) {
	flags := makeFlags()
	flags.Parse(osArgs[1:])
	args := flags.Args()
	if len(args) == 0 {
		return nil, errUsage
	}
	name := args[0]
	c, ok := cmds[name]
	if !ok {
		return nil, fmt.Errorf("no command named %q", name)
	}
	commandWithArgs := func() error {
		return c(args)
	}
	return commandWithArgs, nil
}

var cmds = map[string]command{
	"inputs": commands.Inputs,
}

func makeFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	// Placeholder for when we add flags.
	return fs
}
