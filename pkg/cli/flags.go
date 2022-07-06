package cli

import "flag"

// Flags represents a type that sets options based on
// a set of command line flags.
type Flags interface {
	Flags(*flag.FlagSet)
}

func MergeFlags(flags ...Flags) Flags {
	return mergedFlags(flags)
}

type mergedFlags []Flags

func (mf mergedFlags) Flags(fs *flag.FlagSet) {
	for _, f := range mf {
		f.Flags(fs)
	}
}

func parseFlags(c Command, args []string) ([]string, error) {
	if c.Flags == nil {
		return args, nil
	}
	fs := flag.NewFlagSet(c.Name, flag.ContinueOnError)
	c.Flags.Flags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}
