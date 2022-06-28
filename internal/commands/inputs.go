package commands

import "github.com/hashicorp/actions-go-build/pkg/action"

func Inputs() action.Command {
	return action.Command{
		Subcommands: action.Subcommands{
			"digest": InputsDigest,
		},
	}
}
