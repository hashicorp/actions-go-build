package commands

import (
	"flag"
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

const verifyHelp = `
Verify that a build result is reproducible by attempting to run the same build again.

Args: <build result JSON file>

This command accepts a build result JSON file path, and uses it to run a new verification
build, and compares the results. It downloads a copy of the source code (currently only
supports code hosted on GitHub.com), and uses the config from the result file to run a
similar build in a temporary directory. The new result is compared with the old one, and
a verification result is produced (use the -json flag to print the result to stdout).
`

type verifyOpts struct {
	logOpts
	buildFlags
	present    presenter
	resultFile string
}

func (opts *verifyOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present) }

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	opts.logOpts.Flags(fs)
	opts.buildFlags.ownFlags(fs)
	opts.present.ownFlags(fs)
}

func (opts *verifyOpts) Init() error {
	opts.buildFlags.logOpts = opts.logOpts
	opts.present.logOpts = opts.logOpts
	return nil
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

var Verify = cli.LeafCommand("verify", "verify a build result's reproducibility", func(opts *verifyOpts) error {

	if opts.resultFile == "" {
		return fmt.Errorf("verify requires the -result flag to be set")
	}

	br, err := json.ReadFile[build.Result](opts.resultFile)
	if err != nil {
		return err
	}

	if br.Config.Product.IsDirty() {
		opts.log("WARNING: Result is dirty: source hash != revision")
	}

	sourceURL := fmt.Sprintf("https://github.com/%s/archive/%s.zip", br.Config.Product.Repository, br.Config.Product.Revision)

	b, err := build.NewRemoteVerification(sourceURL, br.Config, opts.buildFlags.buildOpts()...)
	if err != nil {
		return err
	}
	m := opts.buildFlags.newManager(b)

	verifier := verifier.New(br, m, opts.logFunc(), opts.debugFunc())

	result, err := verifier.Verify()
	if err != nil {
		return err
	}

	return opts.present.result("Verification result", result)

}).WithHelp(verifyHelp)
