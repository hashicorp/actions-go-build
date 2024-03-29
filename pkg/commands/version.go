// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type versionOpts struct {
	output
	notrunc bool
	full    bool
	short   bool
}

func (opts *versionOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.output) }

func (opts *versionOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.output)
	fs.BoolVar(&opts.notrunc, "no-trunc", false, "don't truncate the revision SHA")
	fs.BoolVar(&opts.full, "full", false, "print detailed version info")
	fs.BoolVar(&opts.short, "short", false, "print just the unadorned version")
}

// tool is set when MakeVersionCommand is called.
// It is used by other commands, so it's important that MakeVersionCommand
// is called before those commands.
var tool crt.Tool

// MakeVersionCommand makes the version command and returns that along with the default
// version string. This pattern is used so that the main package can inject the version
// info and receive a copy of the default version string that will be returned by this
// command. This is needed to satisfy the --version flag for mitchellh/cli.
func MakeVersionCommand(p crt.Product) (*cli.Command, string) {
	tool = crt.Tool{
		Name:         p.Name,
		Version:      p.Version.Full,
		Revision:     p.Revision,
		RevisionTime: p.RevisionTime,
	}
	trunc := p.VersionCommandOutputShort()
	return cli.LeafCommand("version", "version information", func(opts *versionOpts) error {
		var err error
		if opts.full && opts.short {
			return fmt.Errorf("both -short and -full specified")
		}
		if opts.full {
			if err = opts.output.productInfo(p); err != nil {
				return err
			}
			if p.IsDirty() {
				_, err = fmt.Fprintf(os.Stderr, "Dirty build: SourceHash != Revision\n")
			}
			return err
		}
		if opts.output.json {
			return fmt.Errorf("json output only available when using -full")
		}
		if opts.short {
			_, err = fmt.Fprintln(os.Stdout, p.Version.Full)
			return err
		}
		if opts.notrunc {
			_, err = fmt.Fprintln(os.Stdout, p.VersionCommandOutput())
		} else {
			_, err = fmt.Fprintln(os.Stdout, trunc)
		}
		return err
	}), trunc
}
