package get

import (
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type GitClient struct {
	dir string
}

func NewGitClient(dir string) *GitClient {
	return &GitClient{dir: dir}
}

func Init(dir string) (*git.Repository, error) {
	dotGitFS := osfs.New(filepath.Join(dir, ".git"))
	worktreeFS := osfs.New(dir)
	c := cache.NewObjectLRUDefault()
	s := filesystem.NewStorage(dotGitFS, c)
	return git.Init(s, worktreeFS)
}
