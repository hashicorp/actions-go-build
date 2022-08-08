package commands

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type buildFunc func() (*build.Manager, error)

type configSource int

const (
	configSourceLocal = configSource(iota)
	configSourceRemote
)

type buildish struct {
	desc string
	logOpts
	buildFlags buildFlags
	output     outputOpts

	// target is the only arg
	target string

	// We store these for the sake of verifyish.
	verificationResult *build.VerificationResult
	build              build.Build
	buildResult        *build.Result
	buildConfig        *build.Config
	dir                string
}

// defaultTarget is the default build to target.
// A single '.' means whatever is in the current directory.
const defaultTarget = "."

func (b *buildish) Flags(fs *flag.FlagSet) {
	cli.FlagFuncsAll(fs, b.logOpts.Flags, b.buildFlags.ownFlags, b.output.ownFlags)
}

func (b *buildish) ParseArgs(args []string) error {
	switch len(args) {
	default:
		return fmt.Errorf("at most 1 argument required")
	case 0:
		b.target = defaultTarget
	case 1:
		b.target = args[0]
	}
	return nil
}

func (b *buildish) Init() error {
	b.buildFlags.logOpts = b.logOpts
	b.output.logOpts = b.logOpts
	return nil
}

func (b *buildish) Build(forceVerification bool) (*build.Manager, error) {
	buildFunc, err := b.getBuildFunc(forceVerification)
	if err != nil {
		return nil, err
	}
	return buildFunc()
}

func (b *buildish) runBuild(forceVerification bool) error {
	buildFunc, err := b.getBuildFunc(forceVerification)
	if err != nil {
		return err
	}
	build, err := buildFunc()
	if err != nil {
		return err
	}
	result, err := build.Result()
	if err != nil {
		return err
	}
	return b.output.result(b.desc, result)
}

func (b *buildish) getBuildFunc(forceVerification bool) (buildFunc, error) {
	var (
		done  bool
		err   error
		build buildFunc
	)
	if build, done, err = b.urlConfigSource(b.target); done {
		b.desc = fmt.Sprintf("building using config from %s", b.target)
	} else if build, done, err = b.localDirConfigSource(b.target, forceVerification); done {
		b.desc = fmt.Sprintf("building using config and source code from %s", b.target)
	} else if build, done, err = b.localFileConfigSource(b.target); done {
		b.desc = fmt.Sprintf("building using config from %s", b.target)
	} else {
		err = fmt.Errorf("could not load build config from %q", b.target)
	}
	if err != nil {
		b.debug("error getting build: %s", err)
	}
	b.log(b.desc)
	return build, err
}

func (b *buildish) urlConfigSource(maybeURL string) (buildFunc, bool, error) {
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

func (b *buildish) localFileConfigSource(maybeFile string) (buildFunc, bool, error) {
	maybeFile, exists, err := b.resolvePath("dir", maybeFile, fs.FileExists)
	return b.configSourceFromReadCloser(maybeFile, func() (io.ReadCloser, error) {
		return os.Open(maybeFile)
	}), exists, err
}

func (b *buildish) localDirConfigSource(maybeDir string, forceVerification bool) (buildFunc, bool, error) {
	absDir, exists, err := b.resolvePath("dir", maybeDir, fs.DirExists)
	return func() (*build.Manager, error) {
		c, err := config.FromEnvironment(tool, absDir)
		if err != nil {
			return nil, err
		}
		bc, err := c.PrimaryBuildConfig()
		if err != nil {
			return nil, err
		}
		b.buildConfig = &bc
		b.dir = absDir
		var m *build.Manager
		if forceVerification {
			startTime := time.Now()
			if m, err = b.buildFlags.newLocalVerificationManager(maybeDir, startTime, bc); err != nil {
				return nil, err
			}
		} else if m, err = b.buildFlags.newPrimaryManager(bc); err != nil {
			return nil, err
		}
		b.build = m.Build()
		return m, nil
	}, exists, err
}

func (b *buildish) resolvePath(kind, maybePath string, existsFunc func(string) (bool, error)) (string, bool, error) {
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

func (b *buildish) configSourceFromReadCloser(location string, rcFunc func() (io.ReadCloser, error)) buildFunc {
	return func() (*build.Manager, error) {
		b.debug("reading build config from %q", location)
		rc, err := rcFunc()
		if err != nil {
			return nil, err
		}
		var closeErr error
		defer func() { closeErr = rc.Close() }()
		c, err := b.readConfig(rc)
		if err != nil {
			return nil, fmt.Errorf("unable to read build config from %q: %w", location, err)
		}
		b, err := b.buildFlags.newRemoteVerificationManager(location, c)
		if err != nil {
			return nil, err
		}
		return b, closeErr
	}
}

func (b *buildish) readConfig(r io.Reader) (build.Config, error) {
	if c, ok := tryUnmarshalJSON[build.Config](b.debug, r); ok {
		b.buildConfig = &c
		return c, nil
	}
	if br, ok := tryUnmarshalJSON[build.Result](b.debug, r); ok {
		b.log("Using build config from build result.")
		b.buildResult = &br
		b.buildConfig = &br.Config
		return br.Config, nil
	}
	if vr, ok := tryUnmarshalJSON[build.VerificationResult](b.debug, r); ok {
		b.log("Using primary build config from verification result.")
		b.verificationResult = &vr
		b.buildResult = vr.Primary
		b.buildConfig = &vr.Primary.Config
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
