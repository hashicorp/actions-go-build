package commands

import (
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

type runOpts struct {
	buildFlags
}

var Run = cli.LeafCommand("run", "build a single binary", func(opts *buildFlags) error {
	c, err := opts.buildConfig()
	if err != nil {
		return err
	}
	b, err := build.New(c)
	if err != nil {
		return err
	}
	return b.Run()
})
