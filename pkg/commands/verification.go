package commands

import (
	"fmt"
	"log"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	cp "github.com/otiai10/copy"
)

// Verification runs the verification build, first copying the primary build
// directory to the verification build root.
var Verification = cli.LeafCommand("verification", "run the verification build", func(c *opts.VerificationBuildOpts) error {
	log.Printf("Copying %s to %s", c.PrimaryBuildRoot, c.VerificationBuildRoot)
	if err := cp.Copy(c.PrimaryBuildRoot, c.VerificationBuildRoot); err != nil {
		return err
	}
	log.Printf("Running verification build")
	result := c.Build.Run()
	if _, err := fmt.Fprint(stdout, result); err != nil {
		return err
	}
	return result.Error()
})
