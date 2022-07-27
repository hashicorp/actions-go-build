package commands

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type buildOpts struct {
	Builds       opts.AllBuilds
	Configs      opts.AllBuildConfigs
	ActionConfig opts.ActionConfig
	GitHub       opts.GitHubOpts
	ResultWriter opts.ResultWriter
}

func (bo *buildOpts) ReadEnv() error         { return cli.ReadEnvAll(&bo.Builds, &bo.ActionConfig, &bo.GitHub) }
func (bo *buildOpts) Flags(fs *flag.FlagSet) { cli.FlagsAll(fs, &bo.GitHub) }
func (bo *buildOpts) Init() error            { return cli.InitAll(&bo.Configs) }

var Build = cli.LeafCommand("build", "run primary and local verification build", func(opts *buildOpts) error {

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
			staggerTime, sleepTime)
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

	path, err := opts.ResultWriter.WriteVerificationResult(result)
	if err != nil {
		return err
	}

	if path != "" {
		log.Printf("results written to %s", path)
	}

	return result.Hashes.Error()
})

func primaryBuildResult(opts *buildOpts) (build.Result, error) {
	// See if this build has already been run.
	result, cached, err := opts.Builds.Primary.CachedResult()
	if cached || err != nil {
		log.Printf("Primary build has already been run; skipping.")
		return result, err
	}

	log.Printf("Running primary build.")
	primaryResult := opts.Builds.Primary.Run()
	if primaryResult.Error() != nil {
		if _, err := opts.ResultWriter.WriteBuildResult(primaryResult); err != nil {
			return primaryResult, err
		}
		return primaryResult, fmt.Errorf("primary build failed: %w", primaryResult.Error())
	}
	return primaryResult, nil
}

func verificationBuildResult(opts *buildOpts) (build.Result, error) {
	// See if this build has already been run.
	result, cached, err := opts.Builds.Verification.CachedResult()
	if cached || err != nil {
		log.Printf("Verification build has already been run; skipping.")
		return result, err
	}

	log.Printf("Running verification build.")
	verificationResult, err := runVerificationBuild(
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
	return verificationResult, nil
}
