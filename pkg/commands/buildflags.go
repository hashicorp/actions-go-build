package commands

import (
	"flag"
	"time"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

// buildFlags are flags you can pass to any build, be it primary or verification.
type buildFlags struct {
	logOpts
	rebuild bool
}

func (flags *buildFlags) primaryBuildConfig() (build.Config, error) {
	c, err := config.FromEnvironment(tool)
	if err != nil {
		return build.Config{}, err
	}
	return c.PrimaryBuildConfig()
}

func (flags *buildFlags) localVerificationBuildConfig() (build.Config, error) {
	c, err := config.FromEnvironment(tool)
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

func (flags *buildFlags) newPrimary(c build.Config) (build.Build, error) {
	return build.NewPrimary(c, flags.buildOptions()...)
}

func (flags *buildFlags) newPrimaryManager(c build.Config) (*build.Manager, error) {
	return flags.manager(flags.newPrimary(c))
}

func (flags *buildFlags) newLocalVerification(primaryRoot string, startAfter time.Time, c build.Config) (build.Build, error) {
	return build.NewLocalVerification(primaryRoot, startAfter, c, flags.buildOptions()...)
}

func (flags *buildFlags) newLocalVerificationManager(primaryRoot string, startAfter time.Time, c build.Config) (*build.Manager, error) {
	return flags.manager(flags.newLocalVerification(primaryRoot, startAfter, c))
}

func (flags *buildFlags) newRemoteVerification(sourceURL string, c build.Config) (build.Build, error) {
	return build.NewRemoteVerification(sourceURL, c, flags.buildOptions()...)
}

func (flags *buildFlags) newRemoteVerificationManager(sourceURL string, c build.Config) (*build.Manager, error) {
	return flags.manager(flags.newRemoteVerification(sourceURL, c))
}

func (flags *buildFlags) newVerifier(primary, verification build.ResultSource) (*build.Verifier, error) {
	return build.NewVerifier(primary, verification, flags.buildOptions()...)
}

func (flags *buildFlags) manager(b build.Build, err error) (*build.Manager, error) {
	if err != nil {
		return nil, err
	}
	return flags.newManager(b)
}

func (flags *buildFlags) newManager(b build.Build) (*build.Manager, error) {
	r := build.NewRunner(b, flags.logFunc(), flags.debugFunc())
	return build.NewManager(r, flags.buildOptions()...)
}

func (flags *buildFlags) buildOptions() []build.Option {
	return append(flags.logOpts.buildOptions(), build.WithForceRebuild(flags.rebuild))
}