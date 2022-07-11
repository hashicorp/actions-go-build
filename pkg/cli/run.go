package cli

import "fmt"

func runCLI(c Command, args []string) error {
	subArgs, err := parseFlags(c, args[1:])
	if err != nil {
		return err
	}
	if len(subArgs) == 0 {
		return run(c, nil)
	}
	sub := subArgs[0]
	if len(c.Subcommands()) == 0 {
		return run(c, subArgs)
	}
	sc, ok := getSubCommand(c, sub)
	if !ok {
		return fmt.Errorf("subcommand %q not found", sub)
	}
	return runCLI(sc, subArgs)
}

func run(c Command, args []string) error {
	if c.Run() == nil {
		return ErrNotImplemented
	}
	if a := c.Args(); a != nil {
		a.ParseArgs(args)
	} else if len(args) != 0 {
		return ErrNoArgsAllowed
	}
	return c.Run()()
}
