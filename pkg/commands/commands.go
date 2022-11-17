package commands

import (
	"strings"

	"github.com/mitchellh/cli"

	"github.com/hashicorp/actions-go-build/internal/log"
	actioncli "github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// Commands returns all the commands, optionally prefixed by a set of space-separated
// prefixes (useful when embedding them into other CLI tools).
func Commands(prefix ...string) map[string]cli.CommandFactory {
	p := strings.Join(prefix, " ")
	if len(p) > 0 {
		p += " "
	}
	return map[string]cli.CommandFactory{
		p + "build":   makeCommand(Build),
		p + "config":  makeCommand(Config),
		p + "inspect": makeCommand(Inspect),
		p + "verify":  makeCommand(Verify),
	}
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
