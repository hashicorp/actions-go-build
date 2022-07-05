package crt

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
)

type RepoContext struct {
	RepoName   string
	Dir        string
	CommitSHA  string
	CommitTime time.Time
}

func getRepoName() (string, error) {
	repoName := os.Getenv("PRODUCT_REPOSITORY")
	if repoName != "" {
		return filepath.Base(repoName), nil
	}
	repoName = os.Getenv("GITHUB_REPOSITORY")
	if repoName != "" {
		return filepath.Base(repoName), nil
	}
	return "", fmt.Errorf("Neither GITHUB_REPOSITORY nor PRODUCT_REPOSITORY set")
}

// GetRepoContext reads the repository context from the directory specified.
func GetRepoContext(dir string) (RepoContext, error) {
	repoName, err := getRepoName()
	if err != nil {
		return RepoContext{}, err
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return RepoContext{}, err
	}
	logIter, err := repo.Log(&git.LogOptions{})
	defer logIter.Close()
	commit, err := logIter.Next()
	if err != nil {
		return RepoContext{}, err
	}
	sha := commit.ID().String()
	ts := commit.Author.When

	return RepoContext{
		RepoName:   repoName,
		Dir:        dir,
		CommitSHA:  sha,
		CommitTime: ts,
	}, nil
}
