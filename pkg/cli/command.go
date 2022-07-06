package cli

import (
	"fmt"
	"os"
)

type cmd struct {
	materialise func() Command
	command     *Command
}

func (cf CommandFunc) Command() Command {
	return cf()
}

func (c Command) Command() Command {
	return c
}

type Command struct {
	Name        string
	Run         func() error
	Flags       Flags
	Args        Args
	Synopsis    string
	Subcommands []CommandFunc
}

type CommandFunc func() Command

func Subcommands(commands ...CommandFunc) []CommandFunc {
	return commands
}

// RootCommand is a command that only contains subcommands and doesn't do anything
// by itself.
func RootCommand(name, desc string, subcommands ...CommandFunc) CommandFunc {
	return func() Command {
		run := func() error {
			o := os.Stdout
			p := func(f string, args ...any) { fmt.Fprintf(o, f+"\n", args...) }
			p("%s - %s", name, desc)
			p("")
			p("Subcommands:")
			return TabWrite(o, subcommands, func(cf CommandFunc) string {
				c := cf()
				return fmt.Sprintf("\t%s\t%s", c.Name, c.Synopsis)
			})
		}
		return Command{
			Name:        name,
			Synopsis:    desc,
			Subcommands: subcommands,
			Run:         run,
		}
	}
}

// None can be used as the parameter for leaf commands that have no options.
type None = *any

// LeafCommand is a command that runs something.
// The run function can accept a opts argument, which should be a pointer to a
// type that implements Flags or Args, or both. If no flags or args are needed,
// it can accept None instead.
//
// When the command is run, a new instance of opts (*T) is created first, then its
// flags and args are handled. The resultant opts is passed to the run function.
//
// If opts implements Flags, then its flags are registered with the flag set.
// The Flag set is parsed before the args.
// If opts implements Args, then its ParseArgs method is called on args remaining
// after the flag set is parsed.
// The run function is called after flags and args have been parsed, and passed
// the resultant opts.
func LeafCommand[T any](name, desc string, run func(opts *T) error) CommandFunc {
	return func() Command {
		opts := new(T)
		var (
			flags Flags
			args  Args
		)
		flags, _ = any(opts).(Flags)
		args, _ = any(opts).(Args)
		// Debugging
		//log.Printf("command: %s", name)
		//log.Printf("opts = % #v", opts)
		//log.Printf("*opts = % #v", *opts)
		//log.Printf("flags = % #v", flags)
		//log.Printf("args = % #v", args)
		// End Debugging
		return Command{
			Name:     name,
			Synopsis: desc,
			Flags:    flags,
			Args:     args,
			Run:      func() error { return run(opts) },
		}
	}
}

// Execute executes this command using the provided args.
func (c Command) Execute(args []string) error {
	return runCLI(c, args)
}

func (c Command) getSubCommand(name string) (Command, bool) {
	for _, sc := range c.Subcommands {
		c := sc.Command()
		if c.Name == name {
			return c, true
		}
	}
	return Command{}, false
}
