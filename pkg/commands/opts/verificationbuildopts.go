package opts

import "github.com/hashicorp/actions-go-build/pkg/build"

type VerificationBuildOpts struct {
	Build                 build.Build
	PrimaryBuildRoot      string
	VerificationBuildRoot string
}

func (vbo *VerificationBuildOpts) ReadEnv() error {
	vb := &VerificationBuild{}
	if err := vb.ReadEnv(); err != nil {
		return err
	}
	ac := &ActionConfig{}
	if err := ac.ReadEnv(); err != nil {
		return err
	}
	vbo.Build = vb.Build
	vbo.VerificationBuildRoot = ac.VerificationBuildRoot
	vbo.PrimaryBuildRoot = ac.PrimaryBuildRoot
	return nil
}
