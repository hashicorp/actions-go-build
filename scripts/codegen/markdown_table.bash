# Bash library

markdown_table_rows() {
	column -ts '|' -o ' | ' | sed -E -e 's/^ //g' -e 's/ $//g'
}

markdown_table() {

	HEADER_FUNC="$1"
	BODY_FUNC="$2"
	{ "$HEADER_FUNC"; "$BODY_FUNC"; } | markdown_table_rows
}

column() {
	if [ "$(uname)" = "Darwin" ]; then 
		/usr/local/opt/util-linux/bin/column "$@" || {
			echo "Missing dependency; please install util-linux, e.g.: 'brew install util-linux'"
			return 1
		}
		return
	fi
	column "$@"
}

write_header() {
	write_row "${HEADERS[@]}"
	for _ in "${HEADERS[@]}"; do
		printf "| ----- "
	done
	printf "|\n"
}

write_row() {
	for V in "$@"; do
		printf "| %s " "$V"
	done
	printf "|\n"
}

read_output_to_array() {
	local NAME="$1"; shift
	IFS=$'\n' read -r -d '' -a "$NAME" < <("$@" && printf '\0')
}
