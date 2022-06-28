package commands

import "github.com/hashicorp/actions-go-build/pkg/action"

func Build() action.Command {
	return action.Command{
		Run: func(args []string) error {
			return nil
		},
	}
}
