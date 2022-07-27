package commands

import (
	"fmt"
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

	if err := cacheResult("Verification", result); err != nil {
		return err
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

func cacheResult(name string, result build.Result) error {
	path, err := result.Save()
	if err != nil {
		return fmt.Errorf("Failed to cache build results: %s", err)
	}
	log.Printf("%s build results cached to %s", name, path)
	return nil
}
