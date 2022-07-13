package crt

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
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
	CommitSHA   string
	CommitTime  time.Time
	CoreVersion version.Version
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

	return RepoContext{
		RepoName:    repoName,
		Dir:         dir,
		CommitSHA:   sha,
		CommitTime:  ts,
		CoreVersion: *v,
	}, nil
}

var (
	ErrNoVersionFile        = errors.New("no VERSION file found")
	ErrMultipleVersionFiles = errors.New("multiple VERSION files found")
)

// getCoreVersion exists so that we can add additional version strategies
// in the future. Currently we're only adding a single strategy, which is
// to read from a VERSION file.
func getCoreVersion(dir string) (*version.Version, error) {
	return getCoreVersionFromVersionFile(dir)
}

var versionSearchPath = []string{".", ".release", "dev"}

func getCoreVersionFromVersionFile(dir string) (*version.Version, error) {
	versionFiles, err := findFilesNamed(dir, "VERSION")
	if err != nil {
		return nil, err
	}
	if len(versionFiles) == 0 {
		return nil, ErrNoVersionFile
	}
	if len(versionFiles) > 1 {
		return nil, fmt.Errorf("multiple VERSION files found: %s", strings.Join(versionFiles, ", "))
	}
	vf := versionFiles[0]
	b, err := ioutil.ReadFile(vf)
	if err != nil {
		return nil, err
	}
	vs := strings.TrimSpace(string(b))
	v, err := version.NewVersion(vs)
	if err != nil {
		return nil, fmt.Errorf("parsing version %q from %s: %w", vs, vf, err)
	}
	if m := v.Metadata(); m != "" {
		return nil, fmt.Errorf("version %q contains metadata (from %s)", vs, vf)
	}
	return v, nil
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

func findFilesNamed(dir, name string) ([]string, error) {
	return findFiles(dir, func(d fs.DirEntry, path string) bool {
		return d.Name() == name
	})
}

type findPredicate func(d fs.DirEntry, path string) bool

// findFiles looks for files in the repo, excluding the .git dir.
func findFiles(dir string, predicate findPredicate) ([]string, error) {
	var got []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if predicate(d, path) {
			got = append(got, path)
		}
		return nil
	})
	return got, err
}
