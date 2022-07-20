package opts

import "github.com/hashicorp/actions-go-build/pkg/build"

type EnvDumpOpts struct {
	BuildFlags
	build.Build
}

func (edo *EnvDumpOpts) Init() error {
	if edo.Verification {
		c := &VerificationBuild{}
		if err := c.ReadEnv(); err != nil {
			return err
		}
		edo.Build = c.Build
		return nil
	}
	c := &PrimaryBuild{}
	if err := c.ReadEnv(); err != nil {
		return err
	}
	edo.Build = c.Build
	return nil
}
