package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type inspectOpts struct {
	buildish

	reproducible bool
	goVersion    bool
	buildConfig  bool
	buildEnv     bool
	buildEnvDesc bool
	zipInfo      bool
	worktree     bool
}

func (opts *inspectOpts) Flags(fs *flag.FlagSet) {
	opts.buildish.Flags(fs)
	fs.BoolVar(&opts.buildFlags.forceVerification, "verification", false, "inspect as a verification build")
	fs.BoolVar(&opts.reproducible, "reproducible", false, "just print the reproducible field")
	fs.BoolVar(&opts.goVersion, "go-version", false, "just print the go version")
	fs.BoolVar(&opts.buildConfig, "build-config", false, "just print the build config json")
	fs.BoolVar(&opts.buildEnv, "build-env", false, "just print the build environment")
	fs.BoolVar(&opts.buildEnvDesc, "describe-build-env", false, "describe the build environment")
	fs.BoolVar(&opts.zipInfo, "zip-info", false, "just print the zip details")
	fs.BoolVar(&opts.worktree, "worktree", false, "print worktree status (clean/dirty)")
}

func (opts *inspectOpts) HideFlags() []string {
	return append(opts.logOpts.HideFlags(), "clean", "rebuild", "json")
}

var Inspect = cli.LeafCommand("inspect", "inspect things", func(opts *inspectOpts) error {
	bm, err := opts.build("Inspecting build")
	if err != nil {
		return err
	}

	p := printer{w: os.Stdout, build: bm.Build()}

	if opts.reproducible {
		return p.line("%t", bm.Build().Config().Reproducible)
	}

	if opts.goVersion {
		return p.line(bm.Build().Config().Parameters.GoVersion)
	}

	if opts.buildConfig {
		return json.Write(os.Stdout, bm.Build().Config())
	}

	if opts.buildEnv {
		return p.buildEnv()
	}

	if opts.buildEnvDesc {
		return p.buildEnvDesc()
	}

	if opts.zipInfo {
		return p.zipDetails()
	}

	if opts.worktree {
		return p.worktreeStatus()
	}

	return p.printAll()
})

func (p *printer) printAll() error {
	p.printTitles = true
	p.prefix = "    "
	return firstErr(
		p.worktreeStatus,
		p.buildEnv,
		p.zipDetails,
	)
}

func (p *printer) buildEnv() error {
	if err := p.title("Build Environment"); err != nil {
		return err
	}
	for _, v := range p.build.Env() {
		if err := p.line(v); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) buildEnvDesc() error {
	if err := p.title("Build Environment Description"); err != nil {
		return err
	}
	return tabWrite(p, build.BuildEnvDefinitions(), func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
}

func (p *printer) zipDetails() error {
	if err := p.title("Zip"); err != nil {
		return err
	}
	if err := p.line("ZIP_NAME=%s", p.build.Config().Parameters.ZipName); err != nil {
		return err
	}
	if err := p.line("ZIP_PATH=%s", p.build.Config().Paths.ZipPath); err != nil {
		return err
	}
	return nil
}

func (p *printer) worktreeStatus() error {
	if err := p.title("Worktree Status"); err != nil {
		return err
	}
	product := p.build.Config().Product
	if !product.IsDirty() {
		return p.line("clean")
	}
	if err := p.line("dirty files:"); err != nil {
		return err
	}
	for _, f := range product.DirtyFiles {
		if err := p.line(f); err != nil {
			return err
		}
	}
	return nil
}
