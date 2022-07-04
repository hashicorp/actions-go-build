package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/action"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

func Build() action.Command {
	return action.Command{
		Subcommands: action.Subcommands{
			"run": BuildRun,
			"env": BuildEnv,
		},
	}
}

type buildCommandConfig struct {
	verification bool
}

func (bcc *buildCommandConfig) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("buildrun", flag.ContinueOnError)
	fs.BoolVar(&bcc.verification, "verification", false, "verification build")
	return fs
}

func (bcc *buildCommandConfig) buildConfig() (config.BuildConfig, error) {
	c, err := config.FromEnvironment()
	if err != nil {
		return config.BuildConfig{}, err
	}
	if bcc.verification {
		return c.VerificationBuildConfig()
	}
	return c.PrimaryBuildConfig()
}

func BuildRun() action.Command {
	var bcc buildCommandConfig
	return action.Command{
		Flags: bcc.flagSet,
		Run: func(args []string) error {
			c, err := bcc.buildConfig()
			if err != nil {
				return err
			}
			b, err := build.New(c)
			if err != nil {
				return err
			}
			return b.Run()
		},
	}
}

func BuildEnv() action.Command {
	var bcc buildCommandConfig
	return action.Command{
		Run: func(args []string) error {
			c, err := bcc.buildConfig()
			if err != nil {
				return err
			}
			print(c)
			// TODO: Print build env.
			return nil
		},
	}
}
