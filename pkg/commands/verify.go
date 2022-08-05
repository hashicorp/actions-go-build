package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type verifyOpts struct {
	present    presenter
	resultFile string
	build      buildFlags
}

func (opts *verifyOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present) }

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.present, &opts.build)
	fs.StringVar(&opts.resultFile, "result", "", "path to the json build result file to verify")
}

var Verify = cli.LeafCommand("verify", "verify a build result", func(opts *verifyOpts) error {

	if opts.resultFile == "" {
		return fmt.Errorf("verify requires the -result flag to be set")
	}

	br, err := json.ReadFile[build.Result](opts.resultFile)
	if err != nil {
		return err
	}

	// Update the build paths to a temp dir to run the verification build in.
	tmpDir, err := os.MkdirTemp("", "verification-build.*")
	if err != nil {
		return err
	}
	c, err := br.Config.ChangeRoot(tmpDir)
	if err != nil {
		return err
	}

	b, err := build.New(c)
	if err != nil {
		return err
	}
	vManager := opts.build.newManager(b)

	verifier := verifier.New(br, vManager)

	result, err := verifier.Verify()
	if err != nil {
		return err
	}

	return opts.present.result("Verification result", result)

}).WithHelp(`
Compares the primary and verification build results, and reports an error
if the build did not reproduce correctly.

This command will attempt to use a cached primary build result if no result file is
provideds. It will also use a cached verification build result if available, otherwise
it will run the verification build before doing the comparison.
`)
