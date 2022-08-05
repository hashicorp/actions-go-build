package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type versionOpts struct {
	notrunc bool
}

func (opts *versionOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.notrunc, "no-trunc", false, "don't truncate the revision SHA")
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
		if opts.notrunc {
			_, err = fmt.Fprintln(os.Stdout, p.VersionCommandOutput())
		} else {
			_, err = fmt.Fprintln(os.Stdout, trunc)
		}
		return err
	}), trunc
}
