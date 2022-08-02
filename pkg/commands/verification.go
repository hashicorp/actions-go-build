package commands

import (
	"flag"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

const verificationHelp = `
Run the verification build by making a copy of the current directory in a temporary
path and executing the build instructions there at least 5 seconds later.
` + buildInstructionsHelp

type vBuildOpts struct {
	present presenter
	buildOpts
	staggerTime time.Duration
}

func (opts *vBuildOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present, &opts.buildOpts) }

func (opts *vBuildOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.present, &opts.buildOpts)
	opts.ownFlags(fs)
}

// ownFlags is separated out to make it possible to reuse verification-build-specific flags
// from other places that need to be able to perform verification builds.
func (opts vBuildOpts) ownFlags(fs *flag.FlagSet) {
	fs.DurationVar(&opts.staggerTime, "staggertime", 5*time.Second, "minimum time to wait after start of primary build")
}

// BuildVerification runs the verification build, first copying the primary build
// directory to the verification build root.
var BuildVerification = cli.LeafCommand("verification", "run the verification build", func(opts *vBuildOpts) error {
	vb, err := opts.build()
	if err != nil {
		return err
	}

	result, err := vb.Result()
	if err != nil {
		return err
	}
	return opts.present.result("Verification build", result)
}).WithHelp(verificationHelp)

func (opts *vBuildOpts) build() (*build.Manager, error) { return opts.verificationBuild() }

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
