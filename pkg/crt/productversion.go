package crt

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/go-version"
)

// getCoreVersion exists so that we can add additional version strategies
// in the future. Currently we're only adding a single strategy, which is
// to read from a VERSION file.
func getCoreVersion(dir string) (*version.Version, error) {
	return getCoreVersionFromVersionFile(dir)
}

var versionSearchPath = []string{".", ".release", "dev"}

func versionSearchPaths(basedir string) []string {
	out := make([]string, len(versionSearchPath))
	for i, p := range versionSearchPath {
		out[i] = filepath.Join(basedir, p)
	}
	return out
}

func searchPaths(filename string, paths ...string) (string, error) {
	for _, p := range paths {
		p := filepath.Join(p, filename)
		exists, err := fs.FileExists(p)
		if err != nil {
			return "", err
		}
		if exists {
			return p, nil
		}
	}
	return "", nil
}

func getVersionFile(dir string) (string, error) {
	versionFile, err := searchPaths("VERSION", versionSearchPaths(dir)...)
	if err != nil {
		return "", err
	}
	if len(versionFile) == 0 {
		return "", ErrNoVersionFile
	}
	return versionFile, nil
}

func getCoreVersionFromVersionFile(dir string) (*version.Version, error) {
	versionFile, err := getVersionFile(dir)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return nil, err
	}
	v, err := parseVersion(string(b))
	return v, maybeErr(err, "parsing version file %q", strings.TrimPrefix(versionFile, dir+"/"))
}

func parseVersion(versionString string) (*version.Version, error) {
	vs := strings.TrimSpace(versionString)
	v, err := version.NewVersion(vs)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q", versionString)
	}
	if m := v.Metadata(); m != "" {
		return nil, fmt.Errorf("version %q contains metadata", vs)
	}
	return v, nil
}
