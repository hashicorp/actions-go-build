package commands

import (
	"log"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	cp "github.com/otiai10/copy"
)

// Verification runs the verification build, first copying the primary build
// directory to the verification build root.
var Verification = cli.LeafCommand("verification", "run the verification build", func(c *opts.VerificationBuildOpts) error {
	result, err := runVerificationBuild(c.PrimaryBuildRoot, c.VerificationBuildRoot, c.Build)
	if err != nil {
		return err
	}

	resultFile, err := c.ResultWriter.WriteBuildResult(result)
	if err != nil {
		return err
	}
	if resultFile != "" {
		log.Printf("Verification build results written to %q", resultFile)
	}

	return result.Error()
})

func runVerificationBuild(primaryBuildRoot, verificationBuildRoot string, verificationBuild build.Build) (build.Result, error) {
	log.Printf("Running verification build")
	log.Printf("Copying %s to %s", primaryBuildRoot, verificationBuildRoot)
	if err := cp.Copy(primaryBuildRoot, verificationBuildRoot); err != nil {
		return build.Result{}, err
	}
	return verificationBuild.Run(), nil
}
