package commands

import "github.com/hashicorp/actions-go-build/pkg/cli"

func Main() cli.Command {
	return cli.Command{
		Subcommands: cli.Subcommands{
			"inputs": Inputs,
			"build":  Build,
		},
	}
}
