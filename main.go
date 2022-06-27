package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/actions-go-build/internal/build"
	"github.com/hashicorp/actions-go-build/internal/commands"
)

func main() {
	getCommand(os.Args)
	b := build.New()
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}

var errUsage = fmt.Errorf("usage: %s <command>", os.Args[0])

type command func() error

func getCommand(osArgs []string) (command, error) {
	flags := makeFlags()
	flags.Parse(osArgs[1:])
	args := flags.Args()
	if len(args) != 1 {
		return nil, errUsage
	}
	name := args[0]
	c, ok := cmds[name]
	if !ok {
		return nil, fmt.Errorf("no command named %q", name)
	}
	return c, nil
}

var cmds = map[string]command{
	"digest_inputs": commands.DigestInputs,
}

func makeFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	// Placeholder for when we add flags.
	return fs
}
