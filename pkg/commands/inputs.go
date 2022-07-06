package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

var Inputs = cli.RootCommand("inputs", "parsing build inputs", Digest)

var Digest = cli.LeafCommand("digest", "digest build inputs", func(cli.None) error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	return cfg.ExportToGitHubEnv()
})
