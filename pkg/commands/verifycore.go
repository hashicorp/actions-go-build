package commands

import (
	_ "embed"
	"flag"

	"fmt"
	"io"
	"log"
	"text/template"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/github"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

//go:embed templates/stepsummary.md.tmpl
var stepSummaryTemplate string

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
		log.Printf("Sleeping for %s (%s after initial build start time) to try to trigger temporal nondeterminism.",
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
	if err := writeLogSummary(stderr, result.Hashes); err != nil {
		return err
	}

	path, err := opts.ResultWriter.WriteVerificationResult(result)
	if err != nil {
		return err
	}

	if path != "" {
		log.Printf("results written to %s", path)
	}

	return result.Hashes.Error()
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

func primaryBuildResult(opts *verifyOpts) (build.Result, error) {

	if opts.primaryResultFile != "" {
		r, err := json.ReadFile[build.Result](opts.primaryResultFile)
		if err != nil {
			// Try loading a full verification result instead.
			vr, err := json.ReadFile[build.VerificationResult](opts.primaryResultFile)
			if err != nil {
				return r, err
			}
			if vr.Primary == nil {
				return r, fmt.Errorf("Primary result is nil: %s", opts.primaryResultFile)
			}
			return *vr.Primary, nil
		}
	}

	// See if this build has already been run.
	primaryResult, cached, err := opts.Builds.Primary.CachedResult()
	if cached || err != nil {
		log.Printf("Primary build result found.")
		return primaryResult, err
	}
	if opts.noRunPrimaryBuild {
		return primaryResult, fmt.Errorf("no primary build result found")
	}

	log.Printf("Running primary build.")
	if primaryResult = opts.Builds.Primary.Run(); primaryResult.Error() != nil {
		if _, err := opts.ResultWriter.WriteBuildResult(primaryResult); err != nil {
			return primaryResult, err
		}
		return primaryResult, fmt.Errorf("primary build failed: %w", primaryResult.Error())
	}

	return primaryResult, cacheResult("Primary", primaryResult)
}

func verificationBuildResult(opts *verifyOpts) (build.Result, error) {
	// See if this build has already been run.
	verificationResult, cached, err := opts.Builds.Verification.CachedResult()
	if cached || err != nil {
		log.Printf("Verification build has already been run; skipping.")
		return verificationResult, err
	}
	if opts.noRunVerificationBuild {
		return verificationResult, fmt.Errorf("no verification build result found")
	}

	log.Printf("Running verification build.")
	verificationResult, err = runVerificationBuild(
		opts.ActionConfig.PrimaryBuildRoot,
		opts.ActionConfig.VerificationBuildRoot,
		opts.Builds.Verification,
	)
	if err != nil {
		return verificationResult, fmt.Errorf("setting up for verification build failed: %w", err)
	}
	if verificationResult.Error() != nil {
		if _, err := opts.ResultWriter.WriteBuildResult(verificationResult); err != nil {
			return verificationResult, err
		}
		return verificationResult, fmt.Errorf("verification build failed: %w", verificationResult.Error())
	}
	return verificationResult, cacheResult("Verification", verificationResult)
}
