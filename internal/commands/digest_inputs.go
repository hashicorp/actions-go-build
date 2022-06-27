package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/internal/config"
)

func usageErr(notice string, args []string) error {
	return fmt.Errorf("%s\nusage: %s digest", notice, args[0])
}

func Inputs(args []string) error {
	if len(args) < 2 {
		return usageErr("no arguments provided", args)
	}
	if args[1] != "digest" {
		return usageErr("subcommand not recognised", args)
	}
	return Digest(args[1:])
}

func Digest(args []string) error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}

	cfg.ExportToGitHubEnv()

	return nil
}
