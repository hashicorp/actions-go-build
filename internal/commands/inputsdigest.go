package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/action"
)

func InputsDigest() action.Command {
	return action.Command{
		Run: func(args []string) error {
			cfg, err := config.FromEnvironment()
			if err != nil {
				return err
			}
			cfg.ExportToGitHubEnv()
			return nil
		},
	}
}
