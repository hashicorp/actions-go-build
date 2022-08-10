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

// lvbOpts is "local verification build options"
type lvbOpts struct {
	buildFlags
	logOpts
	output  output
	primary pbOpts

	buildConfig build.Config
	staggerTime time.Duration
}

func (opts *lvbOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.primary)
}

func (opts *lvbOpts) Flags(fs *flag.FlagSet) {
	cli.FlagFuncsAll(fs, opts.logOpts.Flags, opts.primary.ownFlags, opts.output.ownFlags, opts.ownFlags)
}

// ownFlags is separated out to make it possible to reuse verification-build-specific flags
// from other places that need to be able to perform verification builds.
func (opts *lvbOpts) ownFlags(fs *flag.FlagSet) {
	fs.DurationVar(&opts.staggerTime, "staggertime", 5*time.Second, "minimum time to wait after start of primary build")
}

func (opts *lvbOpts) Init() error {
	opts.output.logOpts = opts.logOpts
	opts.primary.logOpts = opts.logOpts
	if err := opts.primary.Init(); err != nil {
		return err
	}
	var err error
	opts.buildConfig, err = opts.buildFlags.localVerificationBuildConfig()
	return err
}

// LVBuild runs the local verification build, first copying the primary build
// directory to the verification build root.
var LVBuild = cli.LeafCommand("local-verification", "run the local verification build", func(opts *lvbOpts) error {
	vb, err := opts.build()
	if err != nil {
		return err
	}

	result, err := vb.Result()
	if err != nil {
		return err
	}
	return opts.output.result("Verification build", result)
}).WithHelp(verificationHelp)

func (opts *lvbOpts) build() (*build.Manager, error) { return opts.verificationBuild() }

func (opts *lvbOpts) verificationBuild() (*build.Manager, error) {
	pb, err := opts.primary.build()
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

	pc := pb.Build().Config()

	return opts.newLocalVerificationManager(pc.Paths.WorkDir, startAfter, opts.buildConfig)
}
