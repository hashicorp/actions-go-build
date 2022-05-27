
# build_env defined which environment variables available to the build_instructions.
# No other environment variables other than PATH are available to the build instructions.
#
# If you set 'PRINT_ENV=true' then this function just writes the variable names and
# their descriptions to stdout. This is useful for generating docs.
build_env() {
	export TARGET_DIR= 

	define_build_env TARGET_DIR "Absolute path to the zip contents directory." "$PWD/$TARGET_DIR"
	
	define_build_env PACKAGE_NAME          "Same as the 'package_name' input."
	define_build_env PRODUCT_VERSION       "Same as the 'product_version' input."
	define_build_env PRODUCT_REVISION      "The git commit SHA of the product repo being built."
	define_build_env PRODUCT_REVISION_TIME "The timestamp of the PRODUCT_REVISION commit in iso-"

	define_build_env BIN_NAME     "Name of the Go binary file inside TARGET_DIR."
	define_build_env BIN_PATH     "Same as TARGET_DIR/BIN_NAME."

	define_build_env OS     "Same as the 'os' input."
	define_build_env ARCH   "Same as the 'arch' input."
	define_build_env GOOS   "Same as OS"
	define_build_env GOARCH "Same as ARCH."

}

print_build_env() {
	PRINT_ENV=true build_env | column -t -s$'\t'
}

define_build_env() {
	local NAME="$1"
	local DESC="$2"
	local SET_TO="${3:-}"

	if [ "$PRINT_ENV" = "true" ]; then
		printf "%s\t%s\n" "$NAME" "$DESC"
		return
	fi

	[ -z "$SET_TO" ] && export "${NAME?}"
	[ -n "$SET_TO" ] && export "${NAME?}"="$SET_TO"
}
