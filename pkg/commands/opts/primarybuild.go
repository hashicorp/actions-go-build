package opts

import (
	"github.com/hashicorp/actions-go-build/pkg/build"
)

type PrimaryBuild struct {
	build.Build
}

func (pb *PrimaryBuild) ReadEnv() error {
	cfg := &PrimaryBuildConfig{}
	if err := cfg.ReadEnv(); err != nil {
		return err
	}
	var err error
	pb.Build, err = build.New(cfg.BuildConfig)
	return err
}
