// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package crt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/git"
	"github.com/hashicorp/go-version"
	"golang.org/x/mod/modfile"
)

type RepoContext struct {
	RepoName    string
	ModuleName  string
	Dir         string
	RootDir     string
	CommitSHA   string
	CommitTime  time.Time
	CoreVersion version.Version
	SourceHash  string
	DirtyFiles  []string `json:",omitempty"`
}

// IsDirty returns true if the worktree is dirty, ignoring
// the dist, out, and meta directories.
func (rc RepoContext) IsDirty() bool {
	return rc.SourceHash == rc.CommitSHA
}

// GetRepoContext reads the repository context from the directory specified.
func GetRepoContext(dir string, ignoreDirs []string) (RepoContext, error) {
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

	worktreeState, err := WorktreeStateFunc(dir, ignoreDirs)
	if err != nil {
		return RepoContext{}, err // blah
	}

	var moduleName string
	goDotMod := filepath.Join(dir, "go.mod")
	exists, err := fs.FileExists(goDotMod)
	if err != nil {
		return RepoContext{}, err
	}
	if exists {
		data, err := ioutil.ReadFile(goDotMod)
		if err != nil {
			return RepoContext{}, err
		}
		m, err := modfile.ParseLax(goDotMod, data, nil)
		if err != nil {
			return RepoContext{}, err
		}
		moduleName = m.Module.Mod.Path
	}

	return RepoContext{
		RepoName:    repoName,
		ModuleName:  moduleName,
		Dir:         dir,
		RootDir:     repo.RootDir(),
		CommitSHA:   sha,
		CommitTime:  ts,
		CoreVersion: *v,
		SourceHash:  worktreeState.SourceHash,
		DirtyFiles:  worktreeState.DirtyFiles,
	}, nil
}

var (
	ErrNoVersionFile        = errors.New("no VERSION file found")
	ErrMultipleVersionFiles = errors.New("multiple VERSION files found")

	// WorktreeStateFunc can be overridden in tests to generate stable
	// source hashes and dirty file lists.
	WorktreeStateFunc = getWorktreeState
)

func getWorktreeState(dir string, ignoreDirs []string) (*git.WorktreeState, error) {
	repo, err := git.Open(dir)
	if err != nil {
		return nil, err
	}
	ignore := makeIgnorePatterns(ignoreDirs)
	s, err := repo.WorktreeState(git.WorktreeStateIgnorePatterns(ignore...))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func makeIgnorePatterns(dirNames []string) []string {
	for i, d := range dirNames {
		dirNames[i] = fmt.Sprintf("^%s\\/", d)
	}
	return dirNames
}

func getRepoName(dir string) (string, error) {
	var repoName string
	if repoName = os.Getenv("PRODUCT_REPOSITORY"); repoName != "" {
		return filepath.Base(repoName), nil
	}
	if repoName = os.Getenv("GITHUB_REPOSITORY"); repoName != "" {
		return filepath.Base(repoName), nil
	}
	// For the sake of running this locally with zero config,
	// we'll guess the repo name by inspecting Git a remote.
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

func getRepoNameFromLocalFilePath(path string) (string, error) {
	path = filepath.ToSlash(path)
	path = strings.TrimSuffix(path, ".")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("Unable to determine repo name from remote %q", path)
	}
	return fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1]), nil
}

func getRepoNameFromRemoteURL(remoteURL string) (string, error) {
	var path string
	if strings.HasPrefix(remoteURL, "/") || strings.HasPrefix(remoteURL, "../") {
		return getRepoNameFromLocalFilePath(remoteURL)
	}
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
