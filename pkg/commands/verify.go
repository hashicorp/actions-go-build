package commands

import "github.com/hashicorp/composite-action-framework-go/pkg/cli"

var Verify = cli.LeafCommand("verify", "verify a build's reproducibility", func(v *verifyish) error {
	return v.runVerification()
})
