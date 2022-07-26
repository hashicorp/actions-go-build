package opts

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

// BuildConfig wraps a crt.BuildConfig to implement the Flags and Env interfaces.
type BuildConfig struct {
	BuildFlags
	build.Config
}

func (bc *BuildConfig) ReadEnv() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	cfgFn := cfg.PrimaryBuildConfig
	if bc.BuildFlags.Verification {
		cfgFn = cfg.VerificationBuildConfig
	}
	bc.Config, err = cfgFn()
	return err
}
