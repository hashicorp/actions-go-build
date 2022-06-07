brew "bats-core"
brew "coreutils"
brew "util-linux"
brew "github-markdown-toc"

# We use force and overwrite to ensure that this parallel
# (the GNU one) is installed instead of the one from coreutils
# which is not usable by bats.
brew "parallel", args: ["force", "overwrite"]
