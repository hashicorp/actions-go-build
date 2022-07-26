package opts

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

type VerificationBuildConfig struct {
	build.Config
}

func (vbc *VerificationBuildConfig) ReadEnv() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	vbc.Config, err = cfg.VerificationBuildConfig()
	return err
}
