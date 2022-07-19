package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// Primary runs the primary build, in the current directory.
var Primary = cli.LeafCommand("primary", "run the primary build", func(cli.None) error {
	return runBuildWithConfig(func(c config.Config) (crt.BuildConfig, error) {
		return c.PrimaryBuildConfig()
	})
})
