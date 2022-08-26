package commands

import (
	"flag"
	"os"
	"time"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

// buildFlags are flags you can pass to any build, be it primary or verification.
type buildFlags struct {
	logOpts
	rebuild bool

	// requireClean and forceVerification are not exposed as flags by default.
	// If a command wants to expose these options it needs to add
	// its own flags to populate these.
	requireClean      bool
	forceVerification bool
}

var wd = func() string {
	w, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return w
}()

func (flags *buildFlags) primaryBuildConfig() (build.Config, error) {
	c, err := config.FromEnvironment(tool, wd)
	if err != nil {
		return build.Config{}, err
	}
	return c.PrimaryBuildConfig()
}

func (flags *buildFlags) localVerificationBuildConfig() (build.Config, error) {
	c, err := config.FromEnvironment(tool, wd)
	if err != nil {
		return build.Config{}, err
	}
	return c.VerificationBuildConfig()
}

func (flags *buildFlags) Flags(fs *flag.FlagSet) {
	flags.logOpts.Flags(fs)
	flags.ownFlags(fs)
}

func (flags *buildFlags) ownFlags(fs *flag.FlagSet) {
	fs.BoolVar(&flags.rebuild, "rebuild", false, "re-run the build even if cached")
}

// A bunch of constructors for things we need configured according to flags.

func (flags *buildFlags) newPrimary(c build.Config, extraOpts ...build.Option) (build.Build, error) {
	return build.NewPrimary(c, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) newPrimaryManager(c build.Config, extraOpts ...build.Option) (*build.Manager, error) {
	p, err := flags.newPrimary(c, extraOpts...)
	return flags.manager(p, err, extraOpts...)
}

func (flags *buildFlags) newRemotePrimary(c build.Config, extraOpts ...build.Option) (build.Build, error) {
	extraOpts = append(extraOpts, build.AsPrimaryBuild())
	return build.NewRemoteBuild(c, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) newRemotePrimaryManager(c build.Config, extraOpts ...build.Option) (*build.Manager, error) {
	p, err := flags.newRemotePrimary(c, extraOpts...)
	return flags.manager(p, err, extraOpts...)
}

func (flags *buildFlags) newLocalVerification(primaryRoot string, startAfter time.Time, c build.Config, extraOpts ...build.Option) (build.Build, error) {
	return build.NewLocalVerification(primaryRoot, startAfter, c, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) newLocalVerificationManager(primaryRoot string, startAfter time.Time, c build.Config, extraOpts ...build.Option) (*build.Manager, error) {
	lv, err := flags.newLocalVerification(primaryRoot, startAfter, c, extraOpts...)
	return flags.manager(lv, err, extraOpts...)
}

func (flags *buildFlags) newRemoteVerification(c build.Config, extraOpts ...build.Option) (build.Build, error) {
	extraOpts = append(extraOpts, build.AsVerificationBuild())
	return build.NewRemoteBuild(c, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) newRemoteVerificationManager(c build.Config, extraOpts ...build.Option) (*build.Manager, error) {
	rv, err := flags.newRemoteVerification(c, extraOpts...)
	return flags.manager(rv, err, extraOpts...)
}

func (flags *buildFlags) newVerifier(primary, verification build.ResultSource, extraOpts ...build.Option) (*build.Verifier, error) {
	return build.NewVerifier(primary, verification, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) manager(b build.Build, err error, extraOpts ...build.Option) (*build.Manager, error) {
	if err != nil {
		return nil, err
	}
	return flags.newManager(b, extraOpts...)
}

func (flags *buildFlags) newManager(b build.Build, extraOpts ...build.Option) (*build.Manager, error) {
	r, err := build.NewRunner(b, flags.buildOptions(extraOpts...)...)
	if err != nil {
		return nil, err
	}
	return build.NewManager(r, flags.buildOptions(extraOpts...)...)
}

func (flags *buildFlags) buildOptions(extraOpts ...build.Option) []build.Option {
	return append(flags.logOpts.buildOptions(extraOpts...),
		build.WithForceRebuild(flags.rebuild),
		build.WithCleanOnly(flags.requireClean),
		build.WithForceVerification(flags.forceVerification),
	)
}
