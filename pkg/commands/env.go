package commands

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/cli"
)

type envOpts struct {
	buildFlags buildFlags
	describe   bool
}

func (ev *envOpts) Flags(fs *flag.FlagSet) {
	ev.buildFlags.Flags(fs)
	fs.BoolVar(&ev.describe, "describe", false, "just describe the flags")
}

var Env = cli.LeafCommand("env", "print the build environment", func(opts *envOpts) error {
	if opts.describe {
		return writeEnvDescriptions()
	}
	return writeEnv(opts.buildFlags)
})

func writeEnv(buildFlags buildFlags) error {
	c, err := buildFlags.buildConfig()
	if err != nil {
		return err
	}
	b, err := build.New(c)
	if err != nil {
		return err
	}
	return tabWrite(b.Env(), func(s string) string { return s })
}

func makeTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
}

func tabWrite[T any](data []T, printer func(datum T) string) error {
	tw := makeTabWriter()
	for _, d := range data {
		s := printer(d) + "\n"
		if _, err := tw.Write([]byte(s)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writeEnvDescriptions() error {
	env := build.BuildEnvDefinitions()
	return tabWrite(env, func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
}
