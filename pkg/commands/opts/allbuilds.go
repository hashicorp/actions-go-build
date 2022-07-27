package opts

import (
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type AllBuilds struct {
	Primary, Verification build.Build

	primary      PrimaryBuild
	verification VerificationBuild
}

func (ab *AllBuilds) ReadEnv() error {
	if err := cli.ReadEnvAll(&ab.primary, &ab.verification); err != nil {
		return err
	}
	ab.Primary = ab.primary.Build
	ab.Verification = ab.verification.Build
	return nil
}
