package opts

import (
	"github.com/hashicorp/actions-go-build/pkg/build"
)

type VerificationBuild struct {
	build.Build
}

func (vb *VerificationBuild) ReadEnv() error {
	cfg := &VerificationBuildConfig{}
	if err := cfg.ReadEnv(); err != nil {
		return err
	}
	var err error
	vb.Build, err = build.New(cfg.Config)
	return err
}
