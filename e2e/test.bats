#!/usr/bin/env bats

set -Eeuo pipefail

# This tests that we get the same verification zip path when we run the commands:
#
# - config
# - inspect -build-config -verification
# - build -verification
@test "verification output paths match" {
	rm -rf dist/ out/
	CONFIG="$(actions-go-build config | grep ZIP_PATH_VERIFICATION | grep -Eo '/.*$')"
	BUILDENV="$(actions-go-build inspect --build-config --verification | jq -r .Paths.ZipPath)"
	actions-go-build build -q
	BUILD="$(actions-go-build build -q -verification -json | jq -r .Config.Paths.ZipPath)"

	echo "$CONFIG"
	echo "$BUILDENV"
	echo "$BUILD"

	diff <(echo "$CONFIG") <(echo "$BUILDENV")
	diff <(echo "$CONFIG") <(echo "$BUILD")

	[[ "$CONFIG" == "$BUILDENV" ]]
}
