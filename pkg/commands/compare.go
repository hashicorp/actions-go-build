package commands

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"text/template"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
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

var Compare = cli.LeafCommand("compare", "compare digests", func(opts *compareOpts) error {
	configs := &opts.AllBuildConfigs
	log.Printf("Comparing SHA256 digests between primary and local verification builds.")
	log.Printf("Primary build root:      %s", configs.Primary.Paths.WorkDir)
	log.Printf("Verification build root: %s", configs.Verification.Paths.WorkDir)
	fsh, err := getAllHashes(&opts.AllBuildConfigs)
	if err != nil {
		return err
	}
	if err := writeStepSummary(opts.StepSummary, fsh); err != nil {
		return err
	}
	return fsh.Error()
})

func writeStepSummary(s github.StepSummary, fsh fileSetHashes) error {
	w, err := s.Open()
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = s.Close() }()

	t := template.Must(template.New("stepsummary").Parse(stepSummaryTemplate))
	if err := t.Execute(w, fsh); err != nil {
		return err
	}

	return closeErr
}

func getAllHashes(configs *opts.AllBuildConfigs) (fileSetHashes, error) {
	getBinPath := func(bc crt.BuildConfig) string { return bc.Paths.BinPath }
	getZipPath := func(bc crt.BuildConfig) string { return bc.Paths.ZipPath }

	var fsh fileSetHashes
	var err error

	if fsh.bin, err = getHashes(configs, getBinPath); err != nil {
		return fsh, err
	}

	if fsh.zip, err = getHashes(configs, getZipPath); err != nil {
		return fsh, err
	}

	return fsh, nil

}

type fileHashes struct {
	primary, verification string
}

// mismatch returns true if the hashes are different, or if they are both empty.
func (fh fileHashes) mismatch() bool {
	return fh.primary != fh.verification && fh.primary != ""
}

type fileSetHashes struct {
	bin fileHashes
	zip fileHashes
}

func (fsh fileSetHashes) Error() error {
	if fsh.bin.mismatch() {
		return fmt.Errorf("executable file mismatch")
	}
	if fsh.zip.mismatch() {
		return fmt.Errorf("zip file mismatch")
	}
	return nil
}

type getPathFunc func(crt.BuildConfig) string

func getHashes(bcs *opts.AllBuildConfigs, getPath func(crt.BuildConfig) string) (fileHashes, error) {
	var fh fileHashes
	var err error
	if fh.primary, err = digest.FileSHA256Hex(getPath(bcs.Primary)); err != nil {
		return fh, err
	}
	if fh.verification, err = digest.FileSHA256Hex(getPath(bcs.Verification)); err != nil {
		return fh, err
	}
	return fh, nil
}
