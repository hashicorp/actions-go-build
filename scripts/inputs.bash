# Bash library

# shellcheck source=scripts/standard_header.bash
source "${BASH_SOURCE%/*}/standard_header.bash"

# This function exports environment variables to GITHUB_ENV
# so they can be used by future steps in the GitHub Actions job.
#
# It either calls 'forward_env' or 'set_env' for each one.
#
#   forward_env: Forwards the already-set value if it's already set
#                or sets it to a default value if provided
#                or results in an error if there's no value or
#   set_env:     Always sets the env to the specified value.
#
digest_inputs() {

	# Pass through env vars from action inputs.

	forward_env PRODUCT_NAME    "$(product_name_from_repo_name "$PRODUCT_REPOSITORY")"

	# We use set_env for PRODUCT_VERSION because even when it's set we want to
	# alter it to add the '+ent; suffix if required.
	set_env     PRODUCT_VERSION "$(apply_ent_version_meta "$PRODUCT_NAME" "$PRODUCT_VERSION")"

	forward_env OS
	forward_env ARCH
	forward_env REPRODUCIBLE
	forward_env INSTRUCTIONS

	# For BIN_NAME we first just forward whatever the user has set (or set it to a default).
	# Then, we set it again, this time ensuring that it has the correct name for the platform.
	forward_env BIN_NAME        "$(remove_enterprise_suffix "$PRODUCT_NAME")"
	set_env     BIN_NAME        "$(ensure_correct_bin_name_for_platform "$BIN_NAME")"

	forward_env ZIP_NAME        "${BIN_NAME}_${PRODUCT_VERSION}_${OS}_${ARCH}.zip"
	
	# Set relative paths used to store various build artifacts.
	
	set_env TARGET_DIR "dist"
	set_env ZIP_DIR    "out"
	set_env META_DIR   ".meta"
	set_env BIN_PATH   "$TARGET_DIR/$BIN_NAME"
	set_env ZIP_PATH   "$ZIP_DIR/$ZIP_NAME"

	# Set absolute paths for the primary and verification builds and artifacts.

	forward_env PRIMARY_BUILD_ROOT "$PWD"
	set_env     BIN_PATH_PRIMARY   "$PRIMARY_BUILD_ROOT/$BIN_PATH"
	set_env     ZIP_PATH_PRIMARY   "$PRIMARY_BUILD_ROOT/$ZIP_PATH"

	forward_env VERIFICATION_BUILD_ROOT "$(adjacent_path "$PWD" "verification")"
	set_env     BIN_PATH_VERIFICATION   "$VERIFICATION_BUILD_ROOT/$BIN_PATH"
	set_env     ZIP_PATH_VERIFICATION   "$VERIFICATION_BUILD_ROOT/$ZIP_PATH"
	
	# Gather contextual info from git.

	set_env PRODUCT_REVISION      "$(git rev-parse HEAD)"
	set_env PRODUCT_REVISION_TIME "$(commit_time_utc "$PRODUCT_REVISION")"

	# Set Go-specific vars.

	set_env GOOS "$OS"
	set_env GOARCH "$ARCH"
}

adjacent_path() { echo "$(dirname "$1")/$2"; }

remove_enterprise_suffix() {
	echo "${1%-enterprise}"
}

ensure_correct_bin_name_for_platform() {
	[[ "$OS" != "windows" ]]    && { echo "$1"; return 0; }
	[[ "$BIN_NAME" = *".exe" ]] && { echo "$1"; return 0; }
	echo "$1.exe"
}

apply_ent_version_meta() {
	local REPO="$1"
	local VERSION="$2"
	trap 'echo "$VERSION"' RETURN
	# If this isn't an enterprise repo, don't make any changes.
	[[ "$REPO" == "$(remove_enterprise_suffix "$REPO")" ]] && return
	# If there's already version metadata, don't make any changes.
	[[ "$VERSION" =~ .*\+.* ]] && return
	# Add the +ent suffix and warn.
	warn "Adding '+ent' to the version because the product_name ends with -enterprise." \
	     "You can remove this warning by adding '+ent' to the version supplied to this action," \
		 "or by changing the product_name to drop the '-enterprise' suffix."
	VERSION="$VERSION+ent"
}

product_name_from_repo_name() {
	basename "$1"
}

commit_time_utc() {
	local COMMIT_ID="$1"
	local T
	T="$(git show -s --format=%cI "$COMMIT_ID")"
	date --utc --iso-8601=seconds -d "$T"
}

export_to_github_job() { local NAME="$1"
	if [ -z "${!NAME+x}" ]; then
		err "$NAME is not set."
		return 1
	fi
	{
		echo "$NAME<<EOF"
		echo "${!NAME}"
		echo "EOF"
	} >> "$GITHUB_ENV"
	# For testing purposes we also write to a standard
	# script file we can source in the tests to see which
	# variables have been exported with which values.
	echo "export $NAME='${!NAME}'" >> "$GITHUB_ENV.export"
	log "Exported to GITHUB_ENV: $NAME='${!NAME}'"
}

# forward_env passes the current value of the named env var
# through to GitHub, or uses the default value if that variable
# is currently empty. If both are empty, it's an error.
forward_env() {
	local NAME="$1"
	local DEFAULT="${2:-}"
	export_env_or_default "$NAME" "$DEFAULT"
	export_to_github_job "$NAME"
}

# set_env sets an env var and preserves it for the 
set_env() {
	local NAME="$1"
	local VALUE="$2"
	export "$NAME"="$VALUE"
	export_env "$NAME"
	export_to_github_job "$NAME"
}

# export_env_or_default exports an env var with the name specified,
# if that variable is already nonempty, then its original value is
# preserved. If it is unset or empty, then it is set to the default
# value.
export_env_or_default() {
	local NAME="$1"
	local DEFAULT="${2:-}"
	# Already got a value? Just export it as that.
	try_export_nonempty "$NAME" && return
	# Default value provided? Export it with that value.
	test -n "$DEFAULT" && {
		info "Using default value for $NAME: '$DEFAULT'"
		export "$NAME"="$DEFAULT"
		return 0
	}
	die "Attempting to export an empty or unset env var '$NAME' with no default value."
}

export_env() {
	local NAME="$1"
	try_export_nonempty "$NAME" || die "Attempting to export an empty or unset env var '$NAME'"
}

try_export_nonempty() {
	local NAME="$1"
	test -n "${!NAME:-}" || return 1
	export "$NAME"="${!NAME}"
}
