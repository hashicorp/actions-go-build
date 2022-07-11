package cli

import "flag"

// Flags represents a type that sets options based on
// a set of command line flags.
type Flags interface {
	Flags(*flag.FlagSet)
}

func parseFlags(c Command, args []string) ([]string, error) {
	f := c.Flags()
	if f == nil {
		return args, nil
	}
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	f.Flags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}
