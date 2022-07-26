package opts

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

type BuildFlags struct {
	Verification bool
}

func (bf *BuildFlags) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&bf.Verification, "verification", false, "verification build")
}

func (bf *BuildFlags) BuildConfig(c config.Config) (build.Config, error) {
	if bf.Verification {
		return c.VerificationBuildConfig()
	}
	return c.PrimaryBuildConfig()
}
