package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type buildFlags struct {
	verification bool
}

func (bcc *buildFlags) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("buildrun", flag.ContinueOnError)
	fs.BoolVar(&bcc.verification, "verification", false, "verification build")
	return fs
}

func (bcc *buildFlags) buildConfig() (crt.BuildConfig, error) {
	c, err := config.FromEnvironment()
	if err != nil {
		return crt.BuildConfig{}, err
	}
	if bcc.verification {
		return c.VerificationBuildConfig()
	}
	return c.PrimaryBuildConfig()
}
