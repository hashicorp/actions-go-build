package get

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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

func GetConfig(dir string) (*config.Config, error) {
	configFile, err := os.Open(filepath.Join(dir, ".git", "config"))
	if err != nil {
		return nil, err
	}
	var closeErr error
	defer func() { closeErr = configFile.Close() }()

	c, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return c, closeErr
}

func GetRemote(dir, name string) (*config.RemoteConfig, error) {
	c, err := GetConfig(dir)
	if err != nil {
		return nil, err
	}
	r, ok := c.Remotes[name]
	if !ok {
		return nil, fmt.Errorf("no remote named %q", name)
	}
	return r, nil
}
