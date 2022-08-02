package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type versionOpts struct {
	notrunc bool
}

func (opts *versionOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.notrunc, "no-trunc", false, "don't truncate the revision SHA")
}

// version and revision are the version and revision of this tool, not of the product
// being built. They are set when MakeVersionCommand is called.
// They are used by other commands, so it's important that MakeVersionCommand
// is called before those commands.
var version, revision string

// MakeVersionCommand makes the version command and returns that along with the default
// version string. This pattern is used so that the main package can inject the version
// info and receive a copy of the default version string that will be returned by this
// command. This is needed to satisfy the --version flag for mitchellh/cli.
func MakeVersionCommand(coreVersion, fullVersion, fullRevision, revisionTime string) (*cli.Command, string) {
	version = versionString(coreVersion, fullVersion)
	revision = fullRevision
	versionTrunc := versionOutput(coreVersion, fullVersion, fullRevision[:8], revisionTime)
	versionNoTrunc := versionOutput(coreVersion, fullVersion, fullRevision, revisionTime)
	return cli.LeafCommand("version", "version information", func(opts *versionOpts) error {
		var err error
		if opts.notrunc {
			_, err = fmt.Fprintln(os.Stdout, versionNoTrunc)
		} else {
			_, err = fmt.Fprintln(os.Stdout, versionTrunc)
		}
		return err
	}), versionTrunc
}

func versionOutput(coreVersion, fullVersion, revision, revisionTime string) string {
	return fmt.Sprintf("v%s %s", versionString(coreVersion, fullVersion), revisionInfo(revision, revisionTime))
}

func versionString(coreVersion, fullVersion string) string {
	if fullVersion != "" {
		return fullVersion
	}
	coreVersion = strings.TrimSpace(coreVersion)
	if coreVersion == "" {
		coreVersion = "0.0.0-unversioned"
	}
	return fmt.Sprintf("%s-local", coreVersion)
}

func revisionInfo(revision, revisionTime string) string {
	if revision == "" {
		return "(unknown revision)"
	}
	revisionString := fmt.Sprintf("(%s)", revision)
	if revisionTime != "" {
		revisionString += fmt.Sprintf(" %s", revisionTime)
	}
	return revisionString
}
