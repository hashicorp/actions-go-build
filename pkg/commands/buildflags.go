package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
)

// buildFlags are flags you can pass to any build, be it primary or verification.
type buildFlags struct {
	logOpts
	rebuild bool
}

func (flags *buildFlags) Flags(fs *flag.FlagSet) {
	flags.logOpts.Flags(fs)
	flags.ownFlags(fs)
}

func (flags *buildFlags) ownFlags(fs *flag.FlagSet) {
	fs.BoolVar(&flags.rebuild, "rebuild", false, "re-run the build even if cached")
}

func (flags *buildFlags) managerOpts() []build.ManagerOption {
	return []build.ManagerOption{
		build.WithLogFunc(flags.logFunc()),
		build.WithForceRebuild(flags.rebuild),
		build.WithDebugLogFunc(flags.debugFunc()),
	}
}

func (flags *buildFlags) newManager(b build.Build) *build.Manager {
	r := build.NewRunner(b, flags.logFunc(), flags.debugFunc())
	return build.NewManager(r, flags.managerOpts()...)
}

func (flags *buildFlags) buildOpts() []build.Option {
	return []build.Option{build.WithLogfunc(flags.logFunc())}
}
