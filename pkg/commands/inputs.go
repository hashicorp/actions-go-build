package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

func Inputs() cli.Command {
	return cli.Command{
		Subcommands: cli.Subcommands{
			"digest": InputsDigest,
		},
	}
}

func InputsDigest() cli.Command {
	return cli.Command{
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
