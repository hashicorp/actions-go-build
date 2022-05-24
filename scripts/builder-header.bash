source scripts/standard-header.bash
source scripts/digest-tools.bash

build() {
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
	)
	
	log "Checking binary artifact written."
	[ -f "$BIN_PATH" ] || die "Binary product $BIN_PATH not found."
	[ -x "$BIN_PATH" ] || die "Binary product $BIN_PATH is not executable."

	write_digest bin "$BIN_PATH"

	log "Zipping contents of '$TARGET_DIR' into '$ZIP_PATH'"
	zip -r -j "$ZIP_PATH" "$TARGET_DIR"

	write_digest zip "$ZIP_PATH"
}

make_and_enter_isolated_build_env() {
	local DIR="$1"
	if [[ "$DIR" != "/"* ]]; then
		die "Directory must be absolute; got '$DIR'"
	fi
	mkdir -p "$DIR"
	cp -r . "$DIR"
	cd "$DIR"
}
