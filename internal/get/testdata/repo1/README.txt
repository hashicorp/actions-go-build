# Test Repo

This repo uses the name 'dotgit' rather than '.git' for its git dir.
This prevents git from treating it as a git repo by default, so we can
make changes and check it in to the larger repo as regular files.

You can make changes to this repo by prefixing git commands with `GIT_DIR=dotgit`.
Run `./edit-repo` to open a shell with this exported by default.
