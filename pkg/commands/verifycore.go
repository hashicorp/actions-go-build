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
	cp "github.com/otiai10/copy"
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
	cli.FlagsAll(fs, &bo.GitHub, &bo.StepSummary, &bo.ResultWriter)
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
		return fmt.Errorf("%w: build ran in %s", err, verificationResult.Config.Paths.WorkDir)
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
		log.Info("RESULT: %s", path)
	}
	if err := result.Hashes.Error(); err != nil {
		return err
	}
	log.Info("OK: Build reproduced correctly.")

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

type buildResultConfig struct {
	name                       string
	build                      build.Build
	rw                         opts.ResultWriter
	loadFromFile               string
	readFromVerificationResult func(build.VerificationResult) *build.Result
	requireExistingBuild       bool
	preBuild                   func() error
}

func buildResult(c buildResultConfig) (build.Result, error) {
	// If a result file is specified, use that.
	if c.loadFromFile != "" {
		r, err := json.ReadFile[build.Result](c.loadFromFile)
		if err != nil {
			// Try loading a full verification result instead.
			vr, err := json.ReadFile[build.VerificationResult](c.loadFromFile)
			if err != nil {
				return r, err
			}
			result := c.readFromVerificationResult(vr)
			if result == nil {
				return r, fmt.Errorf("%s result is nil: %s", c.name, c.loadFromFile)
			}
			r = *result
		}
		log.Info("Primary build result loaded.")
		return r, nil
	}

	// Use the cached result if it exists.
	result, cached, err := c.build.CachedResult()
	if err != nil {
		return result, err
	}
	if cached {
		log.Info("Using cached %s build result.", c.name)
		return result, err
	}
	if c.requireExistingBuild {
		return result, ErrNoPrimaryBuildResult
	}

	// Run the build.
	if c.preBuild != nil {
		if err := c.preBuild(); err != nil {
			return result, err
		}
	}
	log.Info("Running %s build.", c.name)
	if result = c.build.Run(); result.Error() != nil {
		if _, err := c.rw.WriteBuildResult(result); err != nil {
			return result, err
		}
		return result, fmt.Errorf("%s build failed: %w", c.name, result.Error())
	}
	log.Info("OK: Build succeeded: %s", c.name)
	return result, nil
}

func primaryBuildResult(opts *verifyOpts) (build.Result, error) {
	return buildResult(buildResultConfig{
		name:                       "primary",
		build:                      opts.Builds.Primary,
		rw:                         opts.ResultWriter,
		loadFromFile:               opts.primaryResultFile,
		readFromVerificationResult: func(vr build.VerificationResult) *build.Result { return vr.Primary },
		requireExistingBuild:       opts.noRunPrimaryBuild,
		preBuild:                   nil,
	})
}

func verificationBuildResult(opts *verifyOpts) (build.Result, error) {
	return buildResult(buildResultConfig{
		name:                       "verification",
		build:                      opts.Builds.Verification,
		rw:                         opts.ResultWriter,
		loadFromFile:               "",
		readFromVerificationResult: func(vr build.VerificationResult) *build.Result { return vr.Verification },
		requireExistingBuild:       opts.noRunVerificationBuild,
		preBuild: func() error {
			pPath := opts.Builds.Primary.Config().Paths.WorkDir
			vPath := opts.Builds.Verification.Config().Paths.WorkDir
			return cp.Copy(pPath, vPath)
		},
	})
}
