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

	primaryResult, err := runPrimaryBuild(opts)
	if err != nil {
		return err
	}

	sleepTime := 5 * time.Second
	log.Printf("Sleeping for %s to try to trigger temporal nondeterminism.", sleepTime)
	time.Sleep(sleepTime)

	verificationResult, err := doVerificationBuild(opts)
	if err != nil {
		return err
	}

	result, err := build.NewVerificationResult(primaryResult, verificationResult)
	if err != nil {
		return err
	}

	path, err := opts.ResultWriter.WriteDoubleBuildResult(result)
	if err != nil {
		return err
	}

	if path != "" {
		log.Printf("results written to %s", path)
	}

	return result.Hashes.Error()
})

func runPrimaryBuild(opts *buildOpts) (build.Result, error) {
	primaryResult := opts.Builds.Primary.Run()
	if primaryResult.Error() != nil {
		if _, err := opts.ResultWriter.WriteBuildResult(&primaryResult); err != nil {
			return primaryResult, err
		}
		return primaryResult, fmt.Errorf("primary build failed: %w", primaryResult.Error())
	}
	return primaryResult, nil
}

func doVerificationBuild(opts *buildOpts) (build.Result, error) {
	verificationResult, err := runVerificationBuild(
		opts.ActionConfig.PrimaryBuildRoot,
		opts.ActionConfig.VerificationBuildRoot,
		opts.Builds.Verification,
	)
	if err != nil {
		return verificationResult, fmt.Errorf("setting up for verification build failed: %w", err)
	}
	if verificationResult.Error() != nil {
		if _, err := opts.ResultWriter.WriteBuildResult(&verificationResult); err != nil {
			return verificationResult, err
		}
		return verificationResult, fmt.Errorf("verification build failed: %w", verificationResult.Error())
	}
	return verificationResult, nil
}
