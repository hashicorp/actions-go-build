package commands

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type inspectOpts struct {
	buildOpts

	buildEnv bool
	zipName  bool
}

func (opts *inspectOpts) Flags(fs *flag.FlagSet) {
	opts.buildOpts.Flags(fs)
	fs.BoolVar(&opts.buildEnv, "build-env", false, "just print the build environment")
	fs.BoolVar(&opts.zipName, "zip-name", false, "just print the zip name")
}

var Inspect = cli.LeafCommand("inspect", "inspect things", func(opts *inspectOpts) error {
	bm, err := opts.Build("Inspecting build", opts.verification)
	if err != nil {
		return err
	}

	p := printer{w: os.Stdout, build: bm.Build()}

	if opts.buildEnv {
		return p.buildEnv()
	}

	if opts.zipName {
		return p.zipName()
	}

	return p.printAll()
})

type printer struct {
	w           io.Writer
	build       build.Build
	printTitles bool
	prefix      string
}

func (p *printer) printAll() error {
	p.printTitles = true
	p.prefix = "    "
	return firstErr(
		p.buildEnv,
		p.zipName,
	)
}

func (p printer) buildEnv() error {
	if err := p.title("Build Environment"); err != nil {
		return err
	}
	for _, v := range p.build.Env() {
		if err := p.line(p.prefix + v); err != nil {
			return err
		}
	}
	return nil
}

func (p printer) zipName() error {
	if err := p.title("Zip Name"); err != nil {
		return err
	}
	return p.line(p.build.Config().Parameters.ZipName)
}

func firstErr(f ...func() error) error {
	for _, ef := range f {
		if err := ef(); err != nil {
			return err
		}
	}
	return nil
}

func (p printer) title(s string) error {
	if !p.printTitles {
		return nil
	}
	_, err := fmt.Fprintln(p.w, s+":")
	return err
}

func (p printer) line(s string) error {
	_, err := fmt.Fprintln(p.w, s)
	return err
}
