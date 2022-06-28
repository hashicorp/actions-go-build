package commands

import (
	"github.com/hashicorp/actions-go-build/pkg/action"
)

func Main() action.Command {
	return action.Command{
		Subcommands: action.Subcommands{
			"inputs": Inputs,
			"build":  Build,
		},
	}
}
