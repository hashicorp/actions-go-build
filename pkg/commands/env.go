package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/cli"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

var Env = cli.RootCommand("env", "build environment info", EnvDescribe, EnvDump)

var EnvDescribe = cli.LeafCommand("describe", "describe the build environment", func(cli.None) error {
	return writeEnvDescriptions()
})

var EnvDump = cli.LeafCommand("dump", "print the current build environment", func(opts *buildFlags) error {
	c, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	if opts.verification {
		return writeEnv(c.VerificationBuildConfig)
	}
	return writeEnv(c.PrimaryBuildConfig)
})

func writeEnv(bcFunc func() (crt.BuildConfig, error)) error {
	bc, err := bcFunc()
	if err != nil {
		return err
	}
	b, err := build.New(bc)
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
