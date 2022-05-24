write_digest() {
	local NAME="$1"
	local FILE="$2"
	local DIGEST_PATH
	DIGEST_PATH="$(digest_path_rel "$NAME")"
	sha256sum "$FILE" | cut -d' ' > "$DIGEST_PATH" || {
		die "Failed to write digest of '$FILE' to '$DIGEST_PATH'"
	}
	log "Wrote digest of '$FILE' to '$DIGEST_PATH': $(cat "$DIGEST_PATH")"
}

read_digest() {
	local SOURCE_NAME="$1"
	local ROOT_PATH="$1"
	local DIGEST_NAME="$2"
	local DIGEST_PATH="$ROOT_PATH/$META_DIR/${DIGEST_NAME}_digest"
	local DIGEST
	DIGEST="$(cat "$DIGEST_PATH")" || {
		die "Failed to read digest from '$DIGEST_PATH'"
	}
	[ -n "$DIGEST" ] || {
		die "Empty digest read from '$DIGEST_PATH'"
	}
	log "Read $SOURCE_NAME $DIGEST_NAME digest: $DIGEST"
	echo "$DIGEST"
}

digest_names() {
	echo bin
	echo zip
}

digest_path_rel() {
	local DIGEST_NAME="$1"
	digest_names | grep -qF "^$DIGEST_NAME\$" || {
		die "Digest name '$DIGEST_NAME' not recognised; must be one of: $(digest_names | xargs)"
	}
	echo "$META_DIR/${DIGEST_NAME}_digest"	
}

digest_path_abs() {
	local SOURCE_NAME="$1"
	local DIGEST_NAME="$2"
	if [ "$SOURCE_NAME" = "primary" ]; then
		ROOT_PATH="$PRIMARY_ROOT_PATH"
	elif [ "$SOURCE_NAME" = "verification" ]; then
		ROOT_PATH="$LOCAL_VERIFICATION_ROOT_DIR"
	else
		die "Source name '$SOURCE_NAME' not recognised."
	fi
	echo "$ROOT_PATH/$(digest_path_rel "$DIGEST_NAME")"
}

compare_digest() {
	local DIGEST_NAME="$1"

	PRIMARY_DIGEST="$(     read_digest primary      "$DIGEST_NAME")"
	VERIFICATION_DIGEST="$(read_digest verification "$DIGEST_NAME")"

	if [ "$PRIMARY_DIGEST" != "$VERIFICATION_DIGEST" ]; then
		log "FAIL: Digests not equal for $DIGEST_NAME; Primary: $PRIMARY_DIGEST; Verification: $VERIFICATION_DIGEST"
		return 1
	fi
	log "OK: Digests for $DIGEST_NAME are equal: $PRIMARY_BIN_DIGEST"
}
