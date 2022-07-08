package crt

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/actions-go-build/internal/get"
)

type RepoContext struct {
	RepoName   string
	Dir        string
	CommitSHA  string
	CommitTime time.Time
}

// GetRepoContext reads the repository context from the directory specified.
func GetRepoContext(dir string) (RepoContext, error) {
	repoName, err := getRepoName(dir)
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

func getRepoName(dir string) (string, error) {
	repoName := os.Getenv("PRODUCT_REPOSITORY")
	if repoName != "" {
		return filepath.Base(repoName), nil
	}
	repoName = os.Getenv("GITHUB_REPOSITORY")
	if repoName != "" {
		return filepath.Base(repoName), nil
	}
	// For the sake of running this locally, we'll guess
	// the repo name by inspecting Git config.
	origin, err := get.GetRemote(dir, "origin")
	if err == nil && len(origin.URLs) > 0 {
		return getRepoNameFromRemoteURL(origin.URLs[0])
	}

	return "", fmt.Errorf("Neither GITHUB_REPOSITORY nor PRODUCT_REPOSITORY set, and no remote named origin.")
}

func getRepoNameFromRemoteURL(remoteURL string) (string, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(u.Path, ".git"), nil
}
