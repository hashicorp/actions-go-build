package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConfig_ExportToGitHubEnv_ok(t *testing.T) {

	c := standardConfig()

	f, err := os.CreateTemp("", "gh_env.*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	t.Logf("GITHUB_ENV=%q", f.Name())

	os.Setenv("GITHUB_ENV", f.Name())
	c.ExportToGitHubEnv()

	b, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(b), wantGHEnv); diff != "" {
		t.Errorf("ExportToGitHubEnv() mismatch (-want +got):\n%s", diff)
	}
}

const wantGHEnv = `PRODUCT_NAME<<_GitHubActionsFileCommandDelimeter_
lockbox
_GitHubActionsFileCommandDelimeter_
PRODUCT_VERSION<<_GitHubActionsFileCommandDelimeter_
1.2.3
_GitHubActionsFileCommandDelimeter_
PRODUCT_REVISION<<_GitHubActionsFileCommandDelimeter_
cabba9e
_GitHubActionsFileCommandDelimeter_
PRODUCT_REVISION_TIME<<_GitHubActionsFileCommandDelimeter_
2001-12-01T00:00:00Z
_GitHubActionsFileCommandDelimeter_
GO_VERSION<<_GitHubActionsFileCommandDelimeter_
1.18
_GitHubActionsFileCommandDelimeter_
OS<<_GitHubActionsFileCommandDelimeter_
linux
_GitHubActionsFileCommandDelimeter_
ARCH<<_GitHubActionsFileCommandDelimeter_
amd64
_GitHubActionsFileCommandDelimeter_
REPRODUCIBLE<<_GitHubActionsFileCommandDelimeter_
assert
_GitHubActionsFileCommandDelimeter_
INSTRUCTIONS<<_GitHubActionsFileCommandDelimeter_
go build -o $BIN_PATH
_GitHubActionsFileCommandDelimeter_
BIN_NAME<<_GitHubActionsFileCommandDelimeter_
lockbox
_GitHubActionsFileCommandDelimeter_
BIN_PATH<<_GitHubActionsFileCommandDelimeter_
dist/lockbox
_GitHubActionsFileCommandDelimeter_
ZIP_NAME<<_GitHubActionsFileCommandDelimeter_
lockbox_1.2.3_linux_amd64.zip
_GitHubActionsFileCommandDelimeter_
PRIMARY_BUILD_ROOT<<_GitHubActionsFileCommandDelimeter_
/some/dir/work
_GitHubActionsFileCommandDelimeter_
VERIFICATION_BUILD_ROOT<<_GitHubActionsFileCommandDelimeter_
/some/dir/verification
_GitHubActionsFileCommandDelimeter_
PRIMARY_BIN_PATH<<_GitHubActionsFileCommandDelimeter_
/some/dir/work/dist/lockbox
_GitHubActionsFileCommandDelimeter_
PRIMARY_ZIP_PATH<<_GitHubActionsFileCommandDelimeter_
/some/dir/work/out/lockbox_1.2.3_linux_amd64.zip
_GitHubActionsFileCommandDelimeter_
VERIFICATION_BIN_PATH<<_GitHubActionsFileCommandDelimeter_
/some/dir/verification/dist/lockbox
_GitHubActionsFileCommandDelimeter_
VERIFICATION_ZIP_PATH<<_GitHubActionsFileCommandDelimeter_
/some/dir/verification/out/lockbox_1.2.3_linux_amd64.zip
_GitHubActionsFileCommandDelimeter_
TARGET_DIR<<_GitHubActionsFileCommandDelimeter_
dist
_GitHubActionsFileCommandDelimeter_
ZIP_DIR<<_GitHubActionsFileCommandDelimeter_
out
_GitHubActionsFileCommandDelimeter_
META_DIR<<_GitHubActionsFileCommandDelimeter_
meta
_GitHubActionsFileCommandDelimeter_
GOOS<<_GitHubActionsFileCommandDelimeter_
linux
_GitHubActionsFileCommandDelimeter_
GOARCH<<_GitHubActionsFileCommandDelimeter_
amd64
_GitHubActionsFileCommandDelimeter_
`
