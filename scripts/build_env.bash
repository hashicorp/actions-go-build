
# shellcheck disable=SC2016 # I definitely don't want the backticks to be expanded.

# build_env defines which environment variables available to the build_instructions.
#
# If you set 'PRINT_ENV=true' then this function just writes the variable names and
# their descriptions to stdout. This is useful for generating docs.
build_env() {

	if [ "${PRINT_ENV:-}" != "true" ]; then
		make_paths_absolute TARGET_DIR BIN_PATH
	fi

	# The descriptions are markdown formatted.

	define_var TARGET_DIR            'Absolute path to the zip contents directory.'
	define_var PRODUCT_NAME          'Same as the `product_name` input.'
	define_var PRODUCT_VERSION       'Same as the `product_version` input.'
	define_var PRODUCT_REVISION      'The git commit SHA of the product repo being built.'
	define_var PRODUCT_REVISION_TIME 'UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.'
	define_var BIN_NAME              'Name of the Go binary file inside `TARGET_DIR`.'
	define_var BIN_PATH              'Same as `TARGET_DIR/BIN_NAME`.'
	define_var OS                    'Same as the `os` input.'
	define_var ARCH                  'Same as the `arch` input.'
	define_var GOOS                  'Same as `OS`'
	define_var GOARCH                'Same as `ARCH`.'

}

make_paths_absolute() {
	for P in "$@"; do
		[[ -n "${!P:-}" ]] || die "$P is empty"
		export "$P"="$PWD/${!P}"
	done
}

print_build_env() {
	PRINT_ENV=true build_env | column -t -s$'\t'
}

# define_var either exports the named var, or if $PRINT_ENV=true it 
# prints the name and description of each variable in a table.
# This fuction behaves in this way so that we can have a single place
# to define the env vars, and use it both when running the build
# instructions and when generating the corresponding documentation.
define_var() {
	local NAME="$1"
	local DESC="$2"

	if [ "${PRINT_ENV:-}" = "true" ]; then
		printf "%s\t%s\n" "$NAME" "$DESC"
		return
	else
		log "Setting build env: $NAME='${!NAME}'"
	fi

	export "${NAME?}"
}
