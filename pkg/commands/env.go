package commands

import (
	"flag"
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/build"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

var Env = cli.RootCommand("env", "print build environment info", BuildEnvDescribe, BuildEnvDump)

var BuildEnvDescribe = cli.LeafCommand("describe", "describe the build environment", func(cli.None) error {
	env := build.BuildEnvDefinitions()
	return cli.TabWrite(stdout, env, func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
})

type envDumpOpts struct {
	showVerification bool
	primary          pbOpts
	verification     lvbOpts
}

func (opts *envDumpOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.primary, &opts.verification)
}

func (opts *envDumpOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.showVerification, "verification", false, "show the env for a verification build")
}

var BuildEnvDump = cli.LeafCommand("dump", "print the current build environment", func(opts *envDumpOpts) error {
	var b build.Build
	if opts.showVerification {
		m, err := opts.verification.build()
		if err != nil {
			return err
		}
		b = m.Build()
	} else {
		m, err := opts.primary.build()
		if err != nil {
			return err
		}
		b = m.Build()
	}
	return printList(b.Env())
})

func printList(list []string) error {
	return cli.TabWrite(stdout, list, func(s string) string { return s })
}
