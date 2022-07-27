package commands

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"text/template"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/github"
)

//go:embed templates/stepsummary.md.tmpl
var stepSummaryTemplate string

type compareOpts struct {
	opts.AllBuildConfigs
	github.StepSummary
}

func (co *compareOpts) ReadEnv() error {
	return cli.ReadEnvAll(&co.StepSummary, &co.AllBuildConfigs)
}

func (co *compareOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &co.StepSummary)
}

func doComparison(primary, verification build.Config) (crt.FileSetHashes, error) {
	log.Printf("Comparing SHA256 digests between primary and local verification builds.")
	log.Printf("Primary build root:      %s", primary.Paths.WorkDir)
	log.Printf("Verification build root: %s", verification.Paths.WorkDir)
	return build.GetAllHashes(primary, verification)
}

var Compare = cli.LeafCommand("compare", "compare digests", func(opts *compareOpts) error {
	fsh, err := doComparison(opts.AllBuildConfigs.Primary, opts.AllBuildConfigs.Verification)
	if err != nil {
		return err
	}
	if err := writeStepSummary(opts.StepSummary, fsh); err != nil {
		return err
	}
	if err := writeLogSummary(stderr, fsh); err != nil {
		return err
	}
	if err := fsh.Error(); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stderr, "OK: Build reproduced correctly.")
	return err
}).WithHelp(`
Compares the primary and verification build results, and reports an error
if the build did not reproduce correctly.

This command assumes you have already run the 'run primary' and 'run verification'
subcommands to produce the two builds. If you are running this locally, you may
prefer to use the 'build-and-verify' subcommand which ensures those two builds
have been done, and then performs this comparison.
`)

type line struct {
	what, sha string
}

func writeLogSummary(w io.Writer, fsh crt.FileSetHashes) error {
	lines := fileHashLogLines(fsh.Bin, fsh.Zip)
	return cli.TabWrite(w, lines, func(l line) string {
		return fmt.Sprintf("%s\t%s", l.what, l.sha)
	})
}

func fileHashLogLines(fhs ...crt.FileHashes) []line {
	var lines []line
	for _, fh := range fhs {
		lines = append(lines,
			line{fmt.Sprintf("%s - %s", fh.Description, fh.Name), "SHA256"},
			line{"Primary", fh.SHA256.Primary},
			line{"Verification", fh.SHA256.Verification},
			line{},
		)
	}
	return lines
}

func writeStepSummary(s github.StepSummary, fsh crt.FileSetHashes) error {
	w, err := s.Open()
	if err != nil {
		return err
	}
	if w == nil {
		return nil
	}
	var closeErr error
	defer func() { closeErr = s.Close() }()

	t := template.Must(template.New("stepsummary").Parse(stepSummaryTemplate))
	if err := t.Execute(w, fsh); err != nil {
		return err
	}

	return closeErr
}
