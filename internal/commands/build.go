package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/action"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

func Build() action.Command {
	return action.Command{
		Run: func(args []string) error {
			b, err := build.New(config.Config{})
			if err != nil {
				return err
			}
			return b.Run()
		},
	}
}
