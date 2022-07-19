package opts

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type VerificationBuildConfig struct {
	crt.BuildConfig
}

func (vbc *VerificationBuildConfig) ReadEnv() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	vbc.BuildConfig, err = cfg.VerificationBuildConfig()
	return err
}
