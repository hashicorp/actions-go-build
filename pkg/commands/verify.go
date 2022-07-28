package commands

import (
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

var Verify = cli.LeafCommand("verify", "verify a build", func(opts *verifyOpts) error {
	opts.noRunPrimaryBuild = true
	return verifyCore(opts)
}).WithHelp(`
Compares the primary and verification build results, and reports an error
if the build did not reproduce correctly.

This command assumes you have already run the 'run primary' and 'run verification'
subcommands to produce the two builds. If you are running this locally, you may
prefer to use the 'build-and-verify' subcommand which ensures those two builds
have been done, and then performs this comparison.
`)
