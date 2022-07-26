package commands

import (
	_ "embed"
	"flag"
	"log"
	"os"
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
	return co.StepSummary.ReadEnv()
}

func (co *compareOpts) Flags(fs *flag.FlagSet) {
	co.StepSummary.Flags(fs)
}

func (co *compareOpts) Init() error {
	return co.AllBuildConfigs.Init()
}

func doComparison(primary, verification build.BuildConfig) (crt.FileSetHashes, error) {
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
	return fsh.Error()
})

func writeStepSummary(s github.StepSummary, fsh crt.FileSetHashes) error {
	w, err := s.Open()
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stdout
	}
	var closeErr error
	defer func() { closeErr = s.Close() }()

	t := template.Must(template.New("stepsummary").Parse(stepSummaryTemplate))
	if err := t.Execute(w, fsh); err != nil {
		return err
	}

	return closeErr
}
