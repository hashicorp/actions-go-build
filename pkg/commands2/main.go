package commands2

import (
	"log"

	"github.com/hashicorp/actions-go-build/pkg/commands"
	cli2 "github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/mitchellh/cli"
)

var Version string

func MakeCLI(args []string) *cli.CLI {

	c := cli.NewCLI("actions-go-build", Version)

	c.Args = args

	c.Commands = map[string]cli.CommandFactory{
		"build-and-verify":    makeCommand(commands.Verify),
		"run primary":         makeCommand(commands.Primary),
		"run verification":    makeCommand(commands.Verification),
		"verify local":        makeCommand(commands.Compare),
		"config action":       makeCommand(commands.Config),
		"config env describe": makeCommand(commands.EnvDescribe),
		"config env dump":     makeCommand(commands.EnvDump),
	}

	return c
}

type cmd struct {
	help, synopsis string
	run            func([]string) error
}

func (c *cmd) Help() string     { return c.help }
func (c *cmd) Synopsis() string { return c.synopsis }
func (c *cmd) Run(args []string) int {
	if err := c.run(append([]string{""}, args...)); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func makeCommand(command *cli2.Command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &cmd{
			help:     command.Help(),
			synopsis: command.Description(),
			run:      command.Execute,
		}, nil
	}
}
