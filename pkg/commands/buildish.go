package commands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type BuildConfigFunc func() (build.Config, error)

type Buildish struct {
	logOpts
	Config BuildConfigFunc
}

// defaultTarget is the default build to target.
// A single '.' means whatever is in the current directory.
const defaultTarget = "."

func (b *Buildish) ParseArgs(args []string) error {
	switch len(args) {
	default:
		return fmt.Errorf("at most 1 argument required")
	case 0:
		return b.setConfigSource(defaultTarget)
	case 1:
		return b.setConfigSource(args[0])
	}
}

func (b *Buildish) setConfigSource(target string) error {
	var done bool
	var err error
	if b.Config, done, err = b.urlConfigSource(target); done {
		return err
	}
	if b.Config, done, err = b.localFileConfigSource(target); done {
		return err
	}
	if b.Config, done, err = b.localDirConfigSource(target); done {
		return err
	}

	return fmt.Errorf("could not load build config from %q", target)
}

func (b *Buildish) urlConfigSource(maybeURL string) (BuildConfigFunc, bool, error) {
	u, err := url.Parse(maybeURL)
	if err != nil {
		b.debug("not a URL: %s: %s", maybeURL, err)
	}
	if u.Scheme != "https" {
		return nil, false, fmt.Errorf("URLs must use https scheme")
	}
	return b.configSourceFromReadCloser(maybeURL, func() (io.ReadCloser, error) {
		resp, err := http.Get(maybeURL)
		return resp.Body, err
	}), true, err
}

func (b *Buildish) localFileConfigSource(maybeFile string) (BuildConfigFunc, bool, error) {
	maybeFile, exists, err := b.resolvePath("dir", maybeFile, fs.FileExists)
	return b.configSourceFromReadCloser(maybeFile, func() (io.ReadCloser, error) {
		return os.Open(maybeFile)
	}), exists, err
}

func (b *Buildish) localDirConfigSource(maybeDir string) (BuildConfigFunc, bool, error) {
	maybeDir, exists, err := b.resolvePath("dir", maybeDir, fs.DirExists)
	return func() (build.Config, error) {
		c, err := config.FromEnvironment(tool, maybeDir)
		if err != nil {
			return build.Config{}, err
		}
		return c.PrimaryBuildConfig()
	}, exists, err
}

func (b *Buildish) resolvePath(kind, maybePath string, existsFunc func(string) (bool, error)) (string, bool, error) {
	// The dir needs to be absolute, so if it's not, prefix it with the current workdir.
	maybePath, err := ensureAbs(maybePath)
	if err != nil {
		return maybePath, false, err
	}
	exists, err := fs.DirExists(maybePath)
	if err != nil {
		b.loud("unable to check if %s %q exists: %s", kind, err)
	}
	if !exists {
		b.debug("%s doesn't exist: %s", kind, maybePath)
	}
	return maybePath, exists, err
}

func (b *Buildish) configSourceFromReadCloser(location string, rcFunc func() (io.ReadCloser, error)) BuildConfigFunc {
	return func() (build.Config, error) {
		b.debug("reading build config from %q", location)
		rc, err := rcFunc()
		if err != nil {
			return build.Config{}, err
		}
		var closeErr error
		defer func() { closeErr = rc.Close() }()
		c, err := b.readConfig(rc)
		if err != nil {
			return c, fmt.Errorf("unable to read build config from %q: %w", location, err)
		}
		return c, closeErr
	}
}

func (b *Buildish) readConfig(r io.Reader) (build.Config, error) {
	if c, ok := tryUnmarshalJSON[build.Config](b.debug, r); ok {
		return c, nil
	}
	if br, ok := tryUnmarshalJSON[build.Result](b.debug, r); ok {
		b.log("Using build config from build result.")
		return br.Config, nil
	}
	if vr, ok := tryUnmarshalJSON[build.VerificationResult](b.debug, r); ok {
		b.log("Using primary build config from verification result.")
		return vr.Primary.Config, nil
	}
	return build.Config{}, fmt.Errorf("not a build config, build result, or verification result")
}

func tryUnmarshalJSON[T any](debug log.Func, r io.Reader) (T, bool) {
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)
	t := reflect.TypeOf(*(new(T)))
	a, err := json.Read[T](&buf)
	if err != nil {
		debug("not a valid %s: %s", t.Name(), err)
	} else {
		debug("is a valid %s", t.Name())
	}
	return a, err == nil
}

func ensureAbs(maybePath string) (string, error) {
	if filepath.IsAbs(maybePath) {
		return maybePath, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	maybePath = filepath.Join(wd, maybePath)
	return filepath.Clean(maybePath), nil
}
