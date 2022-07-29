package commands

import (
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type verifyOpts struct{}

var Verify = cli.LeafCommand("verify", "verify a build", func(opts *verifyOpts) error {
	return nil
}).WithHelp(`
Compares the primary and verification build results, and reports an error
if the build did not reproduce correctly.

This command will attempt to use a cached primary build result if no result file is
provideds. It will also use a cached verification build result if available, otherwise
it will run the verification build before doing the comparison.
`)
