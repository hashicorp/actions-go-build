source scripts/standard_header.bash
source scripts/digest_tools.bash

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
		# Set TARGET_DIR to absolute in case the instructions cd.
		TARGET_DIR="$(pwd)/$TARGET_DIR" \
			bash "$TEMP_INSTRUCTIONS"
	) || {
		die "Build failed."
	}
	
	log "Checking binary artifact written."
	[ -f "$BIN_PATH" ] || die "Binary product $BIN_PATH not found."
	[ -x "$BIN_PATH" ] || die "Binary product $BIN_PATH is not executable."

	write_digest bin "$BIN_PATH"

	log "Setting created and modified time of all files to be zipped to $PRODUCT_REVISION_TIME_LOCAL"
	for F in "$TARGET_DIR"/*; do
		touch -d "$PRODUCT_REVISION_TIME_LOCAL" "$F"
	done

	log "Zipping contents of '$TARGET_DIR' into '$ZIP_PATH'"
	zip -r -j "$ZIP_PATH" "$TARGET_DIR"

	write_digest zip "$ZIP_PATH"
)}

make_and_enter_isolated_build_env() {
	local DIR="$1"
	if [[ "$DIR" != "/"* ]]; then
		log "Directory must be absolute; got '$DIR'"
		return 1
	fi
	mkdir -p "$DIR"
	cp -r . "$DIR"
	cd "$DIR"
}
