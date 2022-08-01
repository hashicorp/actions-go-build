package commands

import (
	"flag"
	"os"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// BuildVerification runs the verification build, first copying the primary build
// directory to the verification build root.
var BuildVerification = cli.LeafCommand("verification", "run the verification build", func(opts *vBuildOpts) error {
	pb, err := opts.verificationBuild()
	if err != nil {
		return err
	}

	result, err := pb.Result()
	if err != nil {
		return err
	}
	if opts.verbose {
		if err := dumpJSON(os.Stdout, result); err != nil {
			return err
		}
	}

	return result.Error()
}).WithHelp(`
Run the verification build by making a copy of the current directory in a temporary
path and executing the build instructions there at least 5 seconds later.
` + buildInstructionsHelp)

type vBuildOpts struct {
	buildOpts
	staggerTime time.Duration
}

func (opts vBuildOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.buildOpts)
	opts.ownFlags(fs)
}

func (opts vBuildOpts) ownFlags(fs *flag.FlagSet) {
	fs.DurationVar(&opts.staggerTime, "staggertime", 5*time.Second, "minimum time to wait after start of primary build")
}

func (opts *vBuildOpts) verificationBuild() (*build.Manager, error) {
	pb, err := opts.primaryBuild()
	if err != nil {
		return nil, err
	}
	pr, cached, err := pb.ResultFromCache()
	if err != nil {
		return nil, err
	}

	// By default, we'll just wait the staggerTime.
	startAfter := time.Now().Add(opts.staggerTime)
	if cached {
		// If we know when the primary build started we can go a bit quicker.
		startAfter = pr.Meta.Start.Add(opts.staggerTime)
	}

	pc, err := opts.config.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	vc, err := opts.config.VerificationBuildConfig()
	if err != nil {
		return nil, err
	}
	b, err := build.NewLocalVerification(pc.Paths.WorkDir, startAfter, vc, opts.buildOpts.buildOpts()...)
	if err != nil {
		return nil, err
	}
	return opts.newManager(b), nil
}
