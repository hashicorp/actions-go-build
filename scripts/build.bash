

# shellcheck source=scripts/standard_header.bash
source "${BASH_SOURCE%/*}/standard_header.bash"

# shellcheck source=scripts/digest_tools.bash
source "${BASH_SOURCE%/*}/digest_tools.bash"

# shellcheck source=scripts/build_env.bash
source "${BASH_SOURCE%/*}/build_env.bash"

# build performs the build inside a subshell as this allows us
# to exit early with 'die' without killing the caller as well.
build() {(
	set -euo pipefail
	log "Starting build process, rooted at $PWD"

	log "Creating target directory '$TARGET_DIR'"
	mkdir -p "$TARGET_DIR"

	log "Creating zip directory '$ZIP_DIR'"
	mkdir -p "$ZIP_DIR"

	log "Creating metadata directory '$META_DIR'"
	mkdir -p "$META_DIR"

	log "Writing build instructions to temp file..."
	TEMP_INSTRUCTIONS=$(mktemp /tmp/instructions.XXXXXX)
	echo -n "$INSTRUCTIONS" > "$TEMP_INSTRUCTIONS"

	log "Listing the build instructions..."
	{ cat "$TEMP_INSTRUCTIONS"; echo; } 1>&2

	log "Running build instructions..."
	(
			build_env # Set the build env vars.
			bash "$TEMP_INSTRUCTIONS"
	) || {
		die "Build failed."
	}
	
	log "Checking binary artifact written."
	[ -f "$BIN_PATH" ] || die "Binary product $BIN_PATH not found."
	[ -x "$BIN_PATH" ] || die "Binary product $BIN_PATH is not executable."

	write_digest bin "$BIN_PATH"

	local COMMIT_TIME
	COMMIT_TIME="${PRODUCT_REVISION_TIME%+00:00}"
	COMMIT_TIME="${PRODUCT_REVISION_TIME%Z}Z"

	log "Setting created and modified time of all files to be zipped to $COMMIT_TIME"
	for F in "$TARGET_DIR"/*; do
		touch -d "$COMMIT_TIME" "$F"
	done

	log "Zipping contents of '$TARGET_DIR' into '$ZIP_PATH'"
	zip -Xrj "$ZIP_PATH" "$TARGET_DIR"

	write_digest zip "$ZIP_PATH"
)}

make_and_enter_isolated_build_env() {
	local DIR="$1"
	if [[ "$DIR" != "/"* ]]; then
		log "Directory must be absolute; got '$DIR'"
		return 1
	fi
	mkdir -p "$DIR"
	cp -R . "$DIR"
	cd "$DIR"
}
