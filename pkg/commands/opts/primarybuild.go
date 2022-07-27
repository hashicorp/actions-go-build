package opts

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type PrimaryBuild struct {
	build.Build
	ResultWriter
}

func (pb *PrimaryBuild) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &pb.ResultWriter)
}

func (pb *PrimaryBuild) ReadEnv() error {
	if err := cli.ReadEnvAll(&pb.ResultWriter); err != nil {
		return err
	}
	cfg := &PrimaryBuildConfig{}
	if err := cfg.ReadEnv(); err != nil {
		return err
	}
	var err error
	pb.Build, err = build.New(cfg.Config)
	return err
}
