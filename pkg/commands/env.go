package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

var Env = cli.RootCommand("env", "print build environment info", BuildEnvDescribe, BuildEnvDump)

var BuildEnvDescribe = cli.LeafCommand("describe", "describe the build environment", func(cli.None) error {
	return writeEnvDescriptions()
})

var BuildEnvDump = cli.LeafCommand("dump", "print the current build environment", func(opts *opts.EnvDumpOpts) error {
	return printList(opts.Build.Env())
})

func printList(list []string) error {
	return cli.TabWrite(stdout, list, func(s string) string { return s })
}

func writeEnvDescriptions() error {
	env := build.BuildEnvDefinitions()
	return cli.TabWrite(stdout, env, func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
}
