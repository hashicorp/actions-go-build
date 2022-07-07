package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type buildFlags struct {
	verification bool
}

func (bcc *buildFlags) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&bcc.verification, "verification", false, "verification build")
}

func (bcc *buildFlags) buildConfig(c config.Config) (crt.BuildConfig, error) {
	if bcc.verification {
		return c.VerificationBuildConfig()
	}
	return c.PrimaryBuildConfig()
}
