package commands

import (
	"flag"
	"fmt"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type configOpts struct {
	github bool
}

func (c *configOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&c.github, "github", false, "export config to github env")
}

var Config = cli.LeafCommand("config", "print config and export to GITHUB_ENV if set", func(opts *configOpts) error {
	cfg, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	if opts.github {
		return cfg.ExportToGitHubEnv()
	}
	return dumpConfig(cfg)
})

func dumpConfig(c config.Config) error {
	vars, err := c.EnvVars()
	if err != nil {
		return err
	}
	for _, v := range vars {
		fmt.Fprintf(stdout, "%s=%s\n", v.Name, v.Value)
	}
	return nil
}
