package commands

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/hashicorp/actions-go-build/internal/config"
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
	output     output

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

func (b *buildish) Args(args *cli.ArgList) {
	args.Optional(&b.target, "target", ".")
}

func (b *buildish) Init() error {
	b.buildFlags.logOpts = b.logOpts
	b.output.logOpts = b.logOpts
	return nil
}

func (b *buildish) Build(why string, forceVerification bool, extraOpts ...build.Option) (*build.Manager, error) {
	buildFunc, err := b.getBuildFunc(why, forceVerification, extraOpts...)
	if err != nil {
		return nil, err
	}
	return buildFunc()
}

func (b *buildish) runBuild(why string, forceVerification bool) error {
	buildFunc, err := b.getBuildFunc(why, forceVerification)
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

func (b *buildish) getBuildFunc(why string, forceVerification bool, extraOpts ...build.Option) (buildFunc, error) {
	var (
		done  bool
		err   error
		build buildFunc
	)
	b.debug("Resolving buildish %q", b.target)
	b.desc = "Build"
	if build, done, err = b.urlConfigSource(b.target, forceVerification, extraOpts...); done {
		b.log("%s using config from %s", why, b.target)
	} else if build, done, err = b.localDirConfigSource(b.target, forceVerification, extraOpts...); done {
		b.log("%s using config and source code from %s", why, b.target)
	} else if build, done, err = b.localFileConfigSource(b.target, extraOpts...); done {
		b.log("%s using config from %s", why, b.target)
	} else {
		err = fmt.Errorf("could not load build config from %q", b.target)
	}
	if err != nil {
		b.debug("error getting build: %s", err)
	}
	return build, err
}

func (b *buildish) urlConfigSource(maybeURL string, forceVerification bool, extraOpts ...build.Option) (buildFunc, bool, error) {
	u, err := url.Parse(maybeURL)
	if err != nil {
		b.debug("not a URL: %s: %s", maybeURL, err)
	}
	if u.Scheme != "https" {
		return nil, false, fmt.Errorf("URLs must use https scheme")
	}
	return b.configSourceFromReadCloser(maybeURL, forceVerification, func() (io.ReadCloser, error) {
		resp, err := http.Get(maybeURL)
		return resp.Body, err
	}, extraOpts...), true, err
}

func (b *buildish) localFileConfigSource(maybeFile string, extraOpts ...build.Option) (buildFunc, bool, error) {
	maybeFile, exists, err := b.resolvePath("file", maybeFile, fs.FileExists)
	return b.configSourceFromReadCloser(maybeFile, false, func() (io.ReadCloser, error) {
		return os.Open(maybeFile)
	}, extraOpts...), exists, err
}

func (b *buildish) localDirConfigSource(maybeDir string, forceVerification bool, extraOpts ...build.Option) (buildFunc, bool, error) {
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
			if m, err = b.buildFlags.newLocalVerificationManager(maybeDir, startTime, bc, extraOpts...); err != nil {
				return nil, err
			}
		} else if m, err = b.buildFlags.newPrimaryManager(bc, extraOpts...); err != nil {
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
	exists, err := existsFunc(maybePath)
	if err != nil {
		b.loud("unable to check if %s %q exists: %s", kind, err)
	}
	if !exists {
		b.debug("%s doesn't exist: %s", kind, maybePath)
	} else {
		b.debug("%s exists: %s", kind, maybePath)
	}
	return maybePath, exists, err
}

func (b *buildish) configSourceFromReadCloser(location string, forceVerification bool, rcFunc func() (io.ReadCloser, error), extraOpts ...build.Option) buildFunc {
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
		var bm *build.Manager
		if forceVerification {
			bm, err = b.buildFlags.newRemoteVerificationManager(c, extraOpts...)
		} else {
			bm, err = b.buildFlags.newRemotePrimaryManager(c, extraOpts...)
		}
		if err != nil {
			return nil, err
		}
		return bm, closeErr
	}
}

func (b *buildish) readConfig(r io.Reader) (build.Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return build.Config{}, err
	}
	if c, ok := tryUnmarshalJSON[build.Config](b, data); ok {
		b.debug("%s is build config", b.target)
		b.buildConfig = &c
		return c, nil
	}
	if br, ok := tryUnmarshalJSON[build.Result](b, data); ok {
		b.debug("%s is a build result", b.target)
		b.buildResult = &br
		b.buildConfig = &br.Config
		return br.Config, nil
	}
	if vr, ok := tryUnmarshalJSON[build.VerificationResult](b, data); ok {
		b.debug("%s is a verification result", b.target)
		b.verificationResult = &vr
		b.buildResult = vr.Primary
		b.buildConfig = &vr.Primary.Config
		return vr.Primary.Config, nil
	}
	return build.Config{}, fmt.Errorf("not a build config, build result, or verification result")
}

func tryUnmarshalJSON[T any](b *buildish, data []byte) (T, bool) {
	t := reflect.TypeOf(*(new(T)))
	a, err := json.ReadBytes[T](data)
	what := fmt.Sprintf("%s.%s", path.Base(t.PkgPath()), t.Name())
	if err != nil {
		b.debug("%s is not a valid %s: %s", b.target, what, err)
	} else {
		b.debug("%s is is a valid %s", b.target, what)
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
