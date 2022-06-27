package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
)

type RepoContext struct {
	RepoName   string
	WorkDir    string
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

func readRepoContext() (RepoContext, error) {
	repoName, err := getRepoName()
	if err != nil {
		return RepoContext{}, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return RepoContext{}, err
	}
	repo, err := git.PlainOpen(wd)
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
		WorkDir:    wd,
		CommitSHA:  sha,
		CommitTime: ts,
	}, nil
}
