package crt

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/composite-action-framework-go/pkg/git"
	"github.com/hashicorp/go-version"
)

type RepoContext struct {
	RepoName    string
	Dir         string
	RootDir     string
	CommitSHA   string
	CommitTime  time.Time
	CoreVersion version.Version
	SourceHash  string
}

// IsDirty returns true if the worktree is dirty, ignoring
// the dist, out, and meta directories.
func (rc RepoContext) IsDirty() bool {
	return rc.SourceHash == rc.CommitSHA
}

// GetRepoContext reads the repository context from the directory specified.
func GetRepoContext(dir string) (RepoContext, error) {
	repoName, err := getRepoName(dir)
	if err != nil {
		return RepoContext{}, err
	}
	repo, err := git.Open(dir)
	if err != nil {
		return RepoContext{}, err
	}
	commits, err := repo.Log(1)
	if err != nil {
		return RepoContext{}, err
	}
	if len(commits) != 1 {
		return RepoContext{}, fmt.Errorf("no commits")
	}
	sha := commits[0].ID
	ts := commits[0].AuthorTime

	v, err := getCoreVersion(dir)
	if err != nil {
		return RepoContext{}, err
	}

	sourceHash, err := getSourceHash(dir)
	if err != nil {
		return RepoContext{}, err // blah
	}

	return RepoContext{
		RepoName:    repoName,
		Dir:         dir,
		RootDir:     repo.RootDir(),
		CommitSHA:   sha,
		CommitTime:  ts,
		CoreVersion: *v,
		SourceHash:  sourceHash,
	}, nil
}

var (
	ErrNoVersionFile        = errors.New("no VERSION file found")
	ErrMultipleVersionFiles = errors.New("multiple VERSION files found")
)

func getSourceHash(dir string) (string, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return "", err
	}
	ignore := makeIgnorePatterns()
	s, err := repo.WorktreeState(git.WorktreeStateIgnorePatterns(ignore...))
	if err != nil {
		return "", err
	}
	return s.SourceHash, nil
}

func makeIgnorePatterns() []string {
	dirNames := []string{dirs.target, dirs.zip, dirs.meta}
	for i, d := range dirNames {
		dirNames[i] = fmt.Sprintf("^%s\\/", d)
	}
	return dirNames
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
	repo, err := git.Open(dir)
	if err != nil {
		return "", err
	}
	origin, err := repo.GetRemoteNamed("origin")
	if err == nil && len(origin.URLs) > 0 {
		return getRepoNameFromRemoteURL(origin.URLs[0])
	}

	return "", fmt.Errorf("Neither GITHUB_REPOSITORY nor PRODUCT_REPOSITORY set, and no remote named origin.")
}

func getRepoNameFromRemoteURL(remoteURL string) (string, error) {
	var path string
	u, err := url.Parse(remoteURL)
	if err == nil {
		path = u.Path
	} else if path, err = getRepoPathFromRemoteSpecialGitURL(remoteURL); err != nil {
		return "", err
	}
	out := strings.Trim(path, "/")
	return strings.TrimSuffix(out, ".git"), nil
}

func getRepoPathFromRemoteSpecialGitURL(remoteURL string) (string, error) {
	parts := strings.SplitN(remoteURL, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unable to parse remote URL %q", remoteURL)
	}
	return parts[1], nil
}
