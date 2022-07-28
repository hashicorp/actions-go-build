package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	cp "github.com/otiai10/copy"
)

// BuildVerification runs the verification build, first copying the primary build
// directory to the verification build root.
var BuildVerification = cli.LeafCommand("verification", "run the verification build", func(opts *opts.VerificationBuildOpts) error {
	log.Verbose("Running verification build")
	log.Verbose("Copying %s to %s", opts.PrimaryBuildRoot, opts.VerificationBuildRoot)
	result, err := runVerificationBuild(opts.PrimaryBuildRoot, opts.VerificationBuildRoot, opts.Build)
	if err != nil {
		return err
	}

	resultFile, err := opts.ResultWriter.WriteBuildResult(result)
	if err != nil {
		return err
	}
	if resultFile != "" {
		log.Info("Verification build results written to %q", resultFile)
	}

	if err := cacheResult("Verification", result); err != nil {
		return err
	}

	return result.Error()
}).WithHelp(`
Run the verification build by making a copy of the current directory in a temporary
path and executing the build instructions there.
` +
	buildInstructionsHelp)

func runVerificationBuild(primaryBuildRoot, verificationBuildRoot string, verificationBuild build.Build) (build.Result, error) {
	if err := cp.Copy(primaryBuildRoot, verificationBuildRoot); err != nil {
		return build.Result{}, err
	}
	return verificationBuild.Run(), nil
}

func cacheResult(name string, result build.Result) error {
	_, err := result.Save()
	if err != nil {
		return fmt.Errorf("Failed to cache build results: %s", err)
	}
	return nil
}
