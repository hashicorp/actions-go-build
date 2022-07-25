package opts

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
)

type PrimaryBuild struct {
	build.Build
	OutputFile string
}

func (pb *PrimaryBuild) Flags(fs *flag.FlagSet) {
	fs.StringVar(&pb.OutputFile, "output", "", "write build results to the named file")
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
