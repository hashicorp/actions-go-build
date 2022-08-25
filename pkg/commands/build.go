package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type buildOpts struct {
	buildish
	verification bool
	clean        bool
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) {
	opts.buildish.Flags(fs)
	fs.BoolVar(&opts.verification, "verification", false, "configure build as a verification build")
	fs.BoolVar(&opts.clean, "clean", false, "fail unless worktree is clean")
}

var Build = cli.LeafCommand("build", "run a build", func(opts *buildOpts) error {
	build, err := opts.build("Running build", opts.verification, build.WithCleanOnly(opts.clean))
	if err != nil {
		return err
	}
	result, err := build.Result()
	if err != nil {
		return err
	}
	return opts.output.result(opts.desc, result)
})
