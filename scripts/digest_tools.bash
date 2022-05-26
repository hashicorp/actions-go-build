source scripts/standard_header.bash

write_digest() {
	local NAME="$1"
	local FILE="$2"
	local DIGEST_PATH
	DIGEST_PATH="$(digest_path_rel "$NAME")"
	sha256sum "$FILE" | cut -d' ' -f1 > "$DIGEST_PATH" || {
		die "Failed to write digest of '$FILE' to '$DIGEST_PATH'"
	}
	log "Wrote digest of '$FILE' to '$DIGEST_PATH': $(cat "$DIGEST_PATH")"
}

read_digest() {
	local SOURCE_NAME="$1"
	local DIGEST_NAME="$2"
	local DIGEST_PATH
	DIGEST_PATH="$(digest_path_abs "$SOURCE_NAME" "$DIGEST_NAME")"
	local DIGEST
	DIGEST="$(cat "$DIGEST_PATH")" || {
		log "ERROR: Failed to read digest from '$DIGEST_PATH'"
		return 1
	}
	[ -n "$DIGEST" ] || {
		log "ERROR: Empty digest read from '$DIGEST_PATH'"
		return 1
	}
	log "Read $SOURCE_NAME $DIGEST_NAME digest: $DIGEST"
	echo "$DIGEST"
}

assert_digest_name() {
	local DIGEST_NAME="$1"
	local DIGEST_NAMES=(bin zip)
	for N in "${DIGEST_NAMES[@]}"; do
		if [ "$N" = "$DIGEST_NAME" ]; then
			return 0
		fi
	done
	die "Digest name '$DIGEST_NAME' not recognised; must be one of: ${DIGEST_NAMES[*]}"
}

digest_path_rel() {
	local DIGEST_NAME="$1"
	assert_digest_name "$DIGEST_NAME"
	echo "$META_DIR/${DIGEST_NAME}_digest"	
}

digest_path_abs() {
	local SOURCE_NAME="$1"
	local DIGEST_NAME="$2"
	if [ "$SOURCE_NAME" = "primary" ]; then
		ROOT_PATH="$PRIMARY_BUILD_ROOT"
	elif [ "$SOURCE_NAME" = "verification" ]; then
		ROOT_PATH="$VERIFICATION_BUILD_ROOT"
	else
		die "Source name '$SOURCE_NAME' not recognised."
	fi
	echo "$ROOT_PATH/$(digest_path_rel "$DIGEST_NAME")"
}

# compare_digest fails if the two digests are different.
# If it fails, it writes both digests to stdout first primary then verification.
# If it succeeds, it writes just the one digest to stdout.
compare_digest() {
	local DIGEST_NAME="$1"

	PRIMARY_DIGEST="$(     read_digest primary      "$DIGEST_NAME")" || return 1
	VERIFICATION_DIGEST="$(read_digest verification "$DIGEST_NAME")" || return 1

	if [ "$PRIMARY_DIGEST" != "$VERIFICATION_DIGEST" ]; then
		log "FAIL: Digests not equal for $DIGEST_NAME; Primary: $PRIMARY_DIGEST; Verification: $VERIFICATION_DIGEST"
		echo "$PRIMARY_DIGEST"
		echo "$VERIFICATION_DIGEST"
		return 1
	fi
	log "OK: Digests for $DIGEST_NAME are equal: $PRIMARY_DIGEST"
	echo "$PRIMARY_DIGEST"
}
