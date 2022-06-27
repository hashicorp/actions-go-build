package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
)

func DigestInputs() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}

	cfg.ExportToGitHubEnv()

	return nil
}
