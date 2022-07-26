package opts

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

type AllBuildConfigs struct {
	Primary, Verification build.Config
}

func (abc *AllBuildConfigs) Init() error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	if abc.Primary, err = cfg.PrimaryBuildConfig(); err != nil {
		return err
	}
	if abc.Verification, err = cfg.VerificationBuildConfig(); err != nil {
		return err
	}
	return nil
}
