package commands

import (
	"flag"
	"fmt"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type verifyOpts struct {
	present    presenter
	build      buildFlags
	resultFile string
}

func (opts *verifyOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present) }

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.present, &opts.build)
}

func (opts *verifyOpts) ParseArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("argument missing: path to json build result file")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: exactly one required")
	}
	opts.resultFile = args[0]
	return nil
}

var Verify = cli.LeafCommand("verify", "verify a build result", func(opts *verifyOpts) error {

	if opts.resultFile == "" {
		return fmt.Errorf("verify requires the -result flag to be set")
	}

	br, err := json.ReadFile[build.Result](opts.resultFile)
	if err != nil {
		return err
	}

	if br.Config.Product.IsDirty() {
		log.Info("WARNING: Result is dirty: source hash != revision")
	}

	sourceURL := fmt.Sprintf("https://github.com/%s/archive/%s.zip", br.Config.Product.Repository, br.Config.Product.Revision)

	b, err := build.NewRemoteVerification(sourceURL, br.Config, opts.build.buildOpts()...)
	if err != nil {
		return err
	}
	m := opts.build.newManager(b)

	verifier := verifier.New(br, m)

	result, err := verifier.Verify()
	if err != nil {
		return err
	}

	return opts.present.result("Verification result", result)

}).WithHelp(`
Verify that a build result is reproducible.

This command accepts a build result JSON file, uses it to run a new verification
build, and compares the results.
`)
