package config

import (
	"os"
	"testing"

	"github.com/hashicorp/actions-go-build/internal/testhelpers/goldenfile"
)

func TestConfig_ExportToGitHubEnv_ok(t *testing.T) {
	goldenfile.Do(t, func(got *os.File) {
		os.Setenv("GITHUB_ENV", got.Name())
		c := standardConfig()
		c.ExportToGitHubEnv()
	})
}
