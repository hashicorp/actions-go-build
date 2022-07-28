package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	_ "embed"

	"github.com/hashicorp/actions-go-build/pkg/commands"
	actioncli "github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/mitchellh/cli"
)

var (
	//go:embed dev/VERSION
	versionCore                         string
	FullVersion, Revision, RevisionTime string
)

func main() {
	status, err := makeCLI(os.Args[1:], versionOutput()).Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(status)
}

func makeCLI(args []string, version string) *cli.CLI {

	c := cli.NewCLI("actions-go-build", version)

	c.Args = args

	c.Commands = map[string]cli.CommandFactory{
		"":                    makeCommand(commands.BuildAndVerify),
		"build":               makeCommand(commands.Primary),
		"build-and-verify":    makeCommand(commands.BuildAndVerify),
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

func makeCommand(command *actioncli.Command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &cmd{
			help:     command.Help(),
			synopsis: command.Description(),
			run:      command.Execute,
		}, nil
	}
}

func version() string {
	if FullVersion != "" {
		return FullVersion
	}
	versionCore = strings.TrimSpace(versionCore)
	if versionCore == "" {
		versionCore = "0.0.0-unversioned"
	}
	return fmt.Sprintf("%s-local", versionCore)
}

func revision() string {
	if Revision == "" {
		return "(unknown revision)"
	}
	revisionString := fmt.Sprintf("(%s)", Revision[:8])
	if RevisionTime != "" {
		revisionString += fmt.Sprintf(" %s", RevisionTime)
	}
	return revisionString
}

func versionOutput() string {
	return fmt.Sprintf("v%s %s", version(), revision())
}
