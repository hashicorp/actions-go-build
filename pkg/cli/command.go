package cli

import (
	"fmt"
	"os"
)

// cmd represents a command cmd in the CLI graph.
// Don't construct Commands manually, instead use the RootCommand
// and LeafCommand functions to construct root and leaf commands.
type cmd struct {
	name  string
	desc  string
	run   func() error
	flags Flags
	args  Args
	subs  []Command
}

// Command represents a command or subcommand that could be either a leaf
// command (meaning it runs its own custom functionality) or a root command
// (meaning it contains subcommands).
type Command interface {
	Name() string
	Description() string
	Run() func() error
	Flags() Flags
	Args() Args
	Subcommands() []Command
	Execute(args []string) error
}

func (c cmd) Name() string                { return c.name }
func (c cmd) Description() string         { return c.desc }
func (c cmd) Run() func() error           { return c.run }
func (c cmd) Flags() Flags                { return c.flags }
func (c cmd) Args() Args                  { return c.args }
func (c cmd) Subcommands() []Command      { return c.subs }
func (c cmd) Execute(args []string) error { return runCLI(c, args) }

func getSubCommand(parent Command, name string) (Command, bool) {
	for _, c := range parent.Subcommands() {
		if c.Name() == name {
			return c, true
		}
	}
	return cmd{}, false
}

// RootCommand is a command that only contains subcommands and doesn't do anything
// by itself.
func RootCommand(name, desc string, subcommands ...Command) Command {
	run := func() error {
		o := os.Stdout
		p := func(f string, args ...any) { fmt.Fprintf(o, f+"\n", args...) }
		p("%s - %s", name, desc)
		p("")
		p("Subcommands:")
		return TabWrite(o, subcommands, func(c Command) string {
			return fmt.Sprintf("\t%s\t%s", c.Name(), c.Description())
		})
	}
	return cmd{
		name: name,
		desc: desc,
		subs: subcommands,
		run:  run,
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
func LeafCommand[T any](name, desc string, run func(opts *T) error) Command {
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
	return cmd{
		name:  name,
		desc:  desc,
		flags: flags,
		args:  args,
		run:   func() error { return run(opts) },
	}
}
