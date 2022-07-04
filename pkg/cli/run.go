package cli

import "fmt"

func runCLI(cf CommandFunc, args []string) error {
	c := cf()
	subArgs, err := parseFlags(c, args[1:])
	if err != nil {
		return err
	}
	if len(subArgs) == 0 {
		return run(c, nil)
	}
	sub := subArgs[0]
	if len(c.Subcommands) == 0 {
		return run(c, subArgs)
	}
	sc, ok := c.Subcommands[sub]
	if !ok {
		return fmt.Errorf("subcommand %q not found", sub)
	}
	return runCLI(sc, subArgs)
}

func run(c Command, args []string) error {
	if c.Run == nil {
		return ErrNotImplemented
	}
	return c.Run(args)
}

func parseFlags(c Command, args []string) ([]string, error) {
	if c.Flags == nil {
		return args, nil
	}
	fs := c.Flags()
	if fs == nil {
		return args, nil
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}
