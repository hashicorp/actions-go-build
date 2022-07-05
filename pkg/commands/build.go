package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

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

func BuildRun() cli.Command {
	var flags buildFlags
	return cli.Command{
		Flags: flags.flagSet,
		Run: func(args []string) error {
			c, err := flags.buildConfig()
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
	var flags buildFlags
	return cli.Command{
		Flags: flags.flagSet,
		Run: func(args []string) error {
			//c, err := flags.buildConfig()
			//if err != nil {
			//	return err
			//}
			env := build.BuildEnvDefinitions()
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for _, e := range env {
				if _, err := fmt.Fprintf(tw, "%s\t%s\n", e.Name, e.Description); err != nil {
					return err
				}
			}
			tw.Flush()
			return nil
		},
	}
}
