package commands

import (
	_ "embed"
	"errors"
	"flag"

	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/github"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

//go:embed templates/stepsummary.md.tmpl
var stepSummaryTemplate string

var (
	ErrNoPrimaryBuildResult      = errors.New("no primary build result found")
	ErrNoVerificationBuildResult = errors.New("no verification build result found")
)

type verifyOpts struct {
	Builds       opts.AllBuilds
	ActionConfig opts.ActionConfig
	GitHub       opts.GitHubOpts
	StepSummary  github.StepSummary
	ResultWriter opts.ResultWriter

	primaryResultFile string

	// internal opts used for different flavours of verification.
	noRunPrimaryBuild, noRunVerificationBuild bool
}

func (bo *verifyOpts) ReadEnv() error {
	return cli.ReadEnvAll(&bo.Builds, &bo.ActionConfig, &bo.GitHub, &bo.StepSummary)
}

func (bo *verifyOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &bo.GitHub, &bo.StepSummary)
	fs.StringVar(&bo.primaryResultFile, "resultfile", "", "result JSON file to validate (defaults to local cache)")
}

func verifyCore(opts *verifyOpts) error {
	primaryResult, err := primaryBuildResult(opts)
	if err != nil {
		return err
	}

	staggerTime := 5 * time.Second

	earliestVerificationBuildTime := primaryResult.Meta.Start.Add(staggerTime)
	now := time.Now().UTC()
	if earliestVerificationBuildTime.After(now) {
		sleepTime := earliestVerificationBuildTime.Sub(now)
		log.Info("Sleeping for %s (%s after initial build start time) to try to trigger temporal nondeterminism.",
			sleepTime, staggerTime)
		time.Sleep(sleepTime)
	}

	verificationResult, err := verificationBuildResult(opts)
	if err != nil {
		return err
	}

	result, err := build.NewVerificationResult(primaryResult, verificationResult)
	if err != nil {
		return err
	}

	if err := writeStepSummary(opts.StepSummary, result.Hashes); err != nil {
		return err
	}
	if opts.GitHub.GitHubMode {
		if err := writeLogSummary(stderr, result.Hashes); err != nil {
			return err
		}
	}

	path, err := opts.ResultWriter.WriteVerificationResult(result)
	if err != nil {
		return err
	}

	if path != "" {
		log.Info("%s", path)
	}
	if err := result.Hashes.Error(); err != nil {
		return err
	}
	log.Info("Build reproduced correctly.")

	return nil
}

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

func buildResult(name string, b build.Build, rw opts.ResultWriter, loadFromFile string, requireExistingBuild bool) (build.Result, error) {
	// If a result file is specified, use that.
	if loadFromFile != "" {
		r, err := json.ReadFile[build.Result](loadFromFile)
		if err != nil {
			// Try loading a full verification result instead.
			vr, err := json.ReadFile[build.VerificationResult](loadFromFile)
			if err != nil {
				return r, err
			}
			if vr.Primary == nil {
				return r, fmt.Errorf("%s result is nil: %s", name, loadFromFile)
			}
			r = *vr.Primary
		}
		log.Info("Primary build result loaded.")
		return r, nil
	}

	// Use the cached result if it exists.
	primaryResult, cached, err := b.CachedResult()
	if err != nil {
		return primaryResult, err
	}
	if cached {
		log.Info("Using cached %s build result.", name)
		return primaryResult, err
	}
	if requireExistingBuild {
		return primaryResult, ErrNoPrimaryBuildResult
	}

	log.Info("Running %s build.", name)
	if primaryResult = b.Run(); primaryResult.Error() != nil {
		if _, err := rw.WriteBuildResult(primaryResult); err != nil {
			return primaryResult, err
		}
		return primaryResult, fmt.Errorf("%s build failed: %w", name, primaryResult.Error())
	}
	log.Info("OK: Build succeeded: %s", name)
	return primaryResult, nil
}

func primaryBuildResult(opts *verifyOpts) (build.Result, error) {
	return buildResult("primary", opts.Builds.Primary, opts.ResultWriter, opts.primaryResultFile, opts.noRunPrimaryBuild)
}

func verificationBuildResult(opts *verifyOpts) (build.Result, error) {
	return buildResult("verification", opts.Builds.Verification, opts.ResultWriter, "", opts.noRunVerificationBuild)
}
