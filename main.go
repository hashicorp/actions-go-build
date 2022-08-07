package main

import (
	"os"

	_ "embed"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/commands"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/product"
	actioncli "github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/mitchellh/cli"
)

//go:embed dev/VERSION
var versionCore string

func main() {

	p, err := product.Product("actions-go-build", versionCore)
	if err != nil {
		log.Info("Error: invalid build: invalid product info: %s", err)
		os.Exit(1)
	}

	status, err := makeCLI(p, os.Args[1:]).Run()
	if err != nil {
		log.Info("%s", err)
	}
	os.Exit(status)
}

func makeCLI(thisTool crt.Product, args []string) *cli.CLI {

	versionCommand, version := commands.MakeVersionCommand(thisTool)

	c := cli.NewCLI("actions-go-build", version)

	c.Args = args

	c.Commands = map[string]cli.CommandFactory{
		"test":               makeCommand(commands.Test),
		"build primary":      makeCommand(commands.PrimaryBuild),
		"build verification": makeCommand(commands.LVBuild),
		"build env describe": makeCommand(commands.BuildEnvDescribe),
		"build env dump":     makeCommand(commands.BuildEnvDump),
		"verify":             makeCommand(commands.Verify),
		"config":             makeCommand(commands.Config),
		"version":            makeCommand(versionCommand),
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
		log.Info("%s", err)
		return 1
	}
	return 0
}

func makeCommand(command *actioncli.Command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &cmd{
			help:     command.Help(),
			synopsis: command.Description(),
			run:      command.Execute,
		}, nil
	}
}
