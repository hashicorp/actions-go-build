package commands

import (
	"flag"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type buildOpts struct {
	buildish
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) {
	opts.buildish.Flags(fs)
	fs.BoolVar(&opts.buildFlags.forceVerification, "verification", false, "configure build as a verification build")
	fs.BoolVar(&opts.buildFlags.requireClean, "clean", false, "fail unless worktree is clean")
}

var Build = cli.LeafCommand("build", "run a build", func(opts *buildOpts) error {
	build, err := opts.build("Running build")
	if err != nil {
		return err
	}
	result, err := build.Result()
	if err != nil {
		return err
	}
	return opts.output.result(opts.desc, result)
})
