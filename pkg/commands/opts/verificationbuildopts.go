package opts

import (
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type VerificationBuildOpts struct {
	Build                 build.Build
	PrimaryBuildRoot      string
	VerificationBuildRoot string

	ResultWriter

	vbuild VerificationBuild
	config ActionConfig
}

func (vbo *VerificationBuildOpts) ReadEnv() error {
	if err := cli.ReadEnvAll(&vbo.vbuild, &vbo.config); err != nil {
		return err
	}
	vbo.Build = vbo.vbuild.Build
	vbo.VerificationBuildRoot = vbo.config.VerificationBuildRoot
	vbo.PrimaryBuildRoot = vbo.config.PrimaryBuildRoot
	return nil
}
