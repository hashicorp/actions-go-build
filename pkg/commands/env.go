package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

var Env = cli.RootCommand("env", "build environment info", EnvDescribe, EnvDump)

var EnvDescribe = cli.LeafCommand("describe", "describe the build environment", func(cli.None) error {
	return writeEnvDescriptions()
})

var EnvDump = cli.LeafCommand("dump", "print the current build environment", func(opts *buildFlags) error {
	return writeEnv(opts)
})

func writeEnv(buildFlags *buildFlags) error {
	c, err := buildFlags.buildConfig()
	if err != nil {
		return err
	}
	b, err := build.New(c)
	if err != nil {
		return err
	}
	return printList(b.Env())
}

func printList(list []string) error {
	return cli.TabWrite(stdout, list, func(s string) string { return s })
}

func writeEnvDescriptions() error {
	env := build.BuildEnvDefinitions()
	return cli.TabWrite(stdout, env, func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
}
