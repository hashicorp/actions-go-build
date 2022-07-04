package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

func Build() cli.Command {
	return cli.Command{
		Subcommands: cli.Subcommands{
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

func BuildRun() cli.Command {
	var bcc buildCommandConfig
	return cli.Command{
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

func BuildEnv() cli.Command {
	//var bcc buildCommandConfig
	return cli.Command{
		Run: func(args []string) error {
			//c, err := bcc.buildConfig()
			//if err != nil {
			//	return err
			//}
			//print(c)
			//// TODO: Print build env.
			return nil
		},
	}
}
