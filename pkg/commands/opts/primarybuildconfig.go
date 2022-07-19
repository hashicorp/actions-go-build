package opts

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type PrimaryBuildConfig struct {
	crt.BuildConfig
}

func (pbc *PrimaryBuildConfig) ReadEnv() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	pbc.BuildConfig, err = cfg.PrimaryBuildConfig()
	return err
}
