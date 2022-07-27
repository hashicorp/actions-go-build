package commands

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/github"
)

type verifyOpts struct {
	Builds       opts.AllBuilds
	ActionConfig opts.ActionConfig
	GitHub       opts.GitHubOpts
	StepSummary  github.StepSummary
	ResultWriter opts.ResultWriter
}

func (bo *verifyOpts) ReadEnv() error {
	return cli.ReadEnvAll(&bo.Builds, &bo.ActionConfig, &bo.GitHub, &bo.StepSummary)
}

func (bo *verifyOpts) Flags(fs *flag.FlagSet) { cli.FlagsAll(fs, &bo.GitHub, &bo.StepSummary) }

var Verify = cli.LeafCommand("verify", "run primary and verification builds; assert match", func(opts *verifyOpts) error {

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
}).WithHelp(`
Run the primary and verification builds and verify that their outputs are identical.

This command is mostly useful for running locally, when you want to ensure you haven't
introduced accidental nondeterminism into the build. It caches builds, so running it
twice without making any changes to the source code will just use the cached results
on the second and subsequent runs.

Build results are cached according to the SourceHash of the current working directory.
If the working directory is clean (no new or modified files) then the SourceHash is the
same as the HEAD commit SHA. Otherwise it's a hash of that SHA plus the contents of all
new and changed files.
`)

func primaryBuildResult(opts *verifyOpts) (build.Result, error) {
	// See if this build has already been run.
	primaryResult, cached, err := opts.Builds.Primary.CachedResult()
	if cached || err != nil {
		log.Printf("Primary build has already been run; skipping.")
		return primaryResult, err
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
