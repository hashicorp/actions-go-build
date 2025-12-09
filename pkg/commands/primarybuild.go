// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

const primaryBuildHelp = `
Run the primary build by executing the build instructions in the current directory.
` + buildInstructionsHelp

// pbOpts = "primary build opts"
type pbOpts struct {
	logOpts
	buildFlags
	buildConfig build.Config
	output      output
}

func (opts *pbOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.output)
}

func (opts *pbOpts) Flags(fs *flag.FlagSet) {
	cli.FlagFuncsAll(fs, opts.logOpts.Flags, opts.buildFlags.ownFlags, opts.output.ownFlags)
}

// ownFlags does nothing but marks pbOpts as not definining any of its own flags.
func (opts *pbOpts) ownFlags(fs *flag.FlagSet) {}

func (opts *pbOpts) Init() error {
	opts.buildFlags.logOpts = opts.logOpts
	opts.output.logOpts = opts.logOpts
	var err error
	opts.buildConfig, err = opts.primaryBuildConfig()
	return err
}

// PrimaryBuild runs the primary build, in the current directory.
var PrimaryBuild = cli.LeafCommand("primary", "run the primary build", func(opts *pbOpts) error {
	pb, err := opts.build()
	if err != nil {
		return err
	}

	result, err := pb.Result()
	if err != nil {
		return err
	}

	return opts.output.result("Primary build", result)

}).WithHelp(primaryBuildHelp)

func (opts *pbOpts) build() (*build.Manager, error) {
	return opts.buildFlags.newPrimaryManager(opts.buildConfig)
}
