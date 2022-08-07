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

var BuildEnvDump = cli.LeafCommand("dump", "print the current build environment", func(opts *envOpts) error {
	var b build.Build
	if opts.verification {
		m, err := opts.v.build()
		if err != nil {
			return err
		}
		b = m.Build()
	} else {
		m, err := opts.p.build()
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

type envOpts struct {
	verification bool
	p            pbOpts
	v            lvbOpts
}

func (opts *envOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.p, &opts.v)
}

func (opts *envOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.verification, "verification", false, "show the env for the local verification build")
}
