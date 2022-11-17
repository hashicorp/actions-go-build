package main

import (
	"os"
	"strings"

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

// Commands returns all the commands, optionally prefixed by a set of space-separated
// prefixes (useful when embedding them into other CLI tools).
func Commands(prefix ...string) map[string]cli.CommandFactory {
	p := strings.Join(prefix, " ")
	if len(p) > 0 {
		p += " "
	}
	return map[string]cli.CommandFactory{
		p + "build":   makeCommand(commands.Build),
		p + "config":  makeCommand(commands.Config),
		p + "inspect": makeCommand(commands.Inspect),
		p + "verify":  makeCommand(commands.Verify),
	}
}

func makeCLI(thisTool crt.Product, args []string) *cli.CLI {

	versionCommand, version := commands.MakeVersionCommand(thisTool)

	c := cli.NewCLI("actions-go-build", version)

	c.Args = args

	c.Commands = Commands()
	c.Commands["version"] = makeCommand(versionCommand)

	return c
}

type cmd struct {
	*actioncli.Command
	run func([]string) error
}

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
			Command: command,
			run:     command.Execute,
		}, nil
	}
}
