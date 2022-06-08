# bats-core is for running the bats tests
brew "bats-core"

# coreutils is needed for its GNU date program
brew "coreutils"

# util-linux is needed for its column program
# used for generating markdown tables.
brew "util-linux"

# github-markdown-toc is needed for generating
# the readme table of contents.
brew "github-markdown-toc"

# parallel is needed for bats to be able to run
# tests in parallel.
#
# We use force and overwrite to ensure that this parallel
# (the GNU one) is installed instead of the one from moreutils
# which is not usable by bats (in case that is already
# installed).
brew "parallel", args: ["force", "overwrite"]
