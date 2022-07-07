package commands

import (
	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

// runCore just runs the build, it doesn't know if it's a primary or
// verification build.
func runCore(c crt.BuildConfig) error {
	b, err := build.New(c)
	if err != nil {
		return err
	}
	return b.Run()
}

// runBuildWithConfig runs the build defined by the config returned by bcFunc.
// bcFunc is handed the defacto action config from the environment.
func runBuildWithConfig(bcFunc func(config.Config) (crt.BuildConfig, error)) error {
	c, err := config.FromEnvironment()
	if err != nil {
		return err
	}
	bc, err := bcFunc(c)
	if err != nil {
		return err
	}
	return runCore(bc)
}
