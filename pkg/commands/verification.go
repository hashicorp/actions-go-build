package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/cli"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	cp "github.com/otiai10/copy"
)

// Verification runs the verification build, first copying the primary build
// directory to the verification build root.
var Verification = cli.LeafCommand("verification", "run the verification build", func(cli.None) error {
	return runBuildWithConfig(func(c config.Config) (crt.BuildConfig, error) {
		if err := cp.Copy(c.PrimaryBuildRoot, c.VerificationBuildRoot); err != nil {
			return crt.BuildConfig{}, err
		}
		return c.VerificationBuildConfig()
	})
})
