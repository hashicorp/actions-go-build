package opts

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type PrimaryBuild struct {
	build.Build
	ResultWriter

	primary PrimaryBuildConfig
}

func (pb *PrimaryBuild) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &pb.ResultWriter)
}

func (pb *PrimaryBuild) ReadEnv() error {
	if err := cli.ReadEnvAll(&pb.primary); err != nil {
		return err
	}
	var err error
	pb.Build, err = makeBuild(pb.primary.Config)
	return err
}

func makeBuild(c build.Config) (build.Build, error) {
	return build.New(c, build.WithLogfunc(log.Verbose))
}
