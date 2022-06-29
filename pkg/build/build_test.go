package build

import (
	"testing"

	"github.com/hashicorp/actions-go-build/internal/config"
)

func TestBuild_Run_ok(t *testing.T) {

	// Each test case should end up with identical results.

}

func standardConfig(t *testing.T) config.Config {
	return config.Config{
		Inputs: config.Inputs{
			ProductRepository:     "hashicorp/lockbox",
			ProductName:           "lockbox",
			ProductVersion:        "1.2.3",
			GoVersion:             "1.18",
			OS:                    "darwin",
			Arch:                  "arm64",
			Reproducible:          "assert",
			Instructions:          `go build -o "$BIN_PATH" -trimpath -buildvcs=false`,
			BinName:               "lockbox",
			ZipName:               "lockbox_1.2.3_darwin_arm64.zip",
			PrimaryBuildRoot:      "",
			VerificationBuildRoot: "",
		},
		ProductRevision:     "",
		ProductRevisionTime: "",
		ProductCoreName:     "",
		TargetDir:           "",
		ZipDir:              "",
		MetaDir:             "",
	}
}
