package build

import (
	"testing"

	"github.com/hashicorp/actions-go-build/internal/config"
)

func TestBuild_Run_ok(t *testing.T) {

	// Each test case should end up with identical results.

}

func standardConfig(t *testing.T) config.BuildConfig {
	return config.BuildConfig{
		WorkDir:   "/build/root",
		TargetDir: "/dist",
		BinPath:   "/dist/lockbox",
		ZipPath:   "/out/lockbox_1.2.3_amd64.zip",
	}
}
