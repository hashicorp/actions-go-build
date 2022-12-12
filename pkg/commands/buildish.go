// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// buildish represents a thing which can be built. That essentially means anything from which a
// build configuration can be derived. There are currently four things that fit into this category:
//
//   - A local directory coupled with the local system's environment.
//   - A json blob from either a local file or a URL containing:
//     - a build.Config
//     - a build.Result (which contains a build.Config)
//     - a build.VerificationResult (which contains a primary build.Config)
//
// buildish accepts a single argument (target) which can be a local path or a URL.
//
// If target is a local path pointing to a directory, we assume we want to load contextual
// build configuration from that directory and the current environment, and that we want to
// run the build in that same directory, as a primary build. (This can be overridden using
// forceVerification, which copies the current directory to a temporary verification dir
// before running the build there.)
//
// If target is a URL or local path pointing to a config or result file of some sort, then
// we load the build configuration from that file. In this case (when loading config
// from a file) we treat the build as a verification build (i.e. we don't build it in
// the current directory, but rather in a temporary verification build directory).
type buildish struct {
	desc string
	logOpts
	buildFlags buildFlags
	output     output

	// target is the only arg
	target string

	// We store these for the sake of verifyish.
	storedBuild build.Build
	buildResult *build.Result
	buildConfig *build.Config
	dir         string
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

// build is used by consumers of buildish who want a fully-formed build manager which they
// can either execute or inspect. The why parameter is used to make logging more informative.
func (b *buildish) build(why string, extraOpts ...build.Option) (*build.Manager, error) {
	buildFunc, err := b.getBuildFunc(why, extraOpts...)
	if err != nil {
		return nil, err
	}
	return buildFunc()
}

// getBuildFunc uses all the args and flags to return a buildFunc. The strategy is to
// try to intrepret the target in different ways until an interpretation leads to a viable
// build func. The order we try this interpretation probably doesn't matter for correctness.
//
// Currently the order is:
//
//   1. URL pointing to build config.
//   2. Local directory containing source code.
//   3. Local file containing build config.
//
// Note that we don't eagerly try to actually read the build config, because options on the
// buildish may still be further tweaked by calling code to influence the settings of the
// build, so we only lazily get the build config itself at the last minute when it's needed.
func (b *buildish) getBuildFunc(why string, extraOpts ...build.Option) (buildFunc, error) {

	var (
		// done is set to true when we've found the correct interpretation of target.
		done bool
		// err is the error returned from generating the buildFunc
		err error
		// build is the buildFunc itself that we want
		build buildFunc
	)

	b.debug("Resolving buildish %q", b.target)
	b.desc = "Build"

	if build, done, err = b.urlConfigSource(b.target, extraOpts...); done {

		// Target is a URL, hopefully pointing to a JSON blob containing build config.
		b.log("%s using config from %s", why, b.target)

	} else if build, done, err = b.localDirConfigSource(b.target, extraOpts...); done {

		// Target is a local directory containing source code to be built.
		absTarget, err := filepath.Abs(b.target)
		if err != nil {
			return nil, err
		}
		b.log("%s using config and source code from %s", why, absTarget)

	} else if build, done, err = b.localFileConfigSource(b.target, extraOpts...); done {

		// Target is a local file, hopefully containing a JSON blob containing build config.
		b.log("%s using config from %s", why, b.target)

	} else {

		// Target is gobbledygook.
		err = fmt.Errorf("could not load build config from %q", b.target)
	}

	if err != nil {
		b.debug("error getting build: %s", err)
	}

	return build, err
}

// localFileConfigSource returns a buildFunc which derives build config from a JSON blob retrieved via HTTPS.
func (b *buildish) urlConfigSource(maybeURL string, extraOpts ...build.Option) (buildFunc, bool, error) {
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
	}, extraOpts...), true, err
}

// localFileConfigSource returns a buildFunc which derives build config from a JSON blob stored in a local file.
func (b *buildish) localFileConfigSource(maybeFile string, extraOpts ...build.Option) (buildFunc, bool, error) {
	maybeFile, exists, err := b.resolvePath("file", maybeFile, fs.FileExists)
	return b.configSourceFromReadCloser(maybeFile, func() (io.ReadCloser, error) {
		return os.Open(maybeFile)
	}, extraOpts...), exists, err
}

// localDirConfigSource returns a buildFunc which derives build config from source code in a local
// directory, alongside the current environment.
func (b *buildish) localDirConfigSource(maybeDir string, extraOpts ...build.Option) (buildFunc, bool, error) {
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
		if b.buildFlags.forceVerification {
			startTime := time.Now()
			if m, err = b.buildFlags.newLocalVerificationManager(maybeDir, startTime, bc, extraOpts...); err != nil {
				return nil, err
			}
		} else if m, err = b.buildFlags.newPrimaryManager(bc, extraOpts...); err != nil {
			return nil, err
		}
		b.storedBuild = m.Build()
		return m, nil
	}, exists, err
}

// resolvePath returns the absolute version of maybePath, alongside a boolean indicating
// if that path passes the existsFunc test. The kind parameter is used to make logging richer.
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

// configSourceFromReadCloser accepts a function (rcFunc) that obtains an io.ReadCloser, and creates a
// buildFunc that inteprets bytes from that io.ReadCloser as a json blob containing build configuration.
// We use rcFunc rather than an actual io.ReadCloser so that we don't need to open files or make requests
// until the last possible moment when they're needed. This avoids eagerly loading data we don't end up
// needing.
//
// The location parameter is just used for logging purposes, and is assumed to indicate a file path or URL
// from which the readcloser is sourced.
func (b *buildish) configSourceFromReadCloser(location string, rcFunc func() (io.ReadCloser, error), extraOpts ...build.Option) buildFunc {
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

		if c.Product.IsDirty() {
			return nil, fmt.Errorf("unable to run a remote build based on a dirty build result")
		}

		var bm *build.Manager
		if b.buildFlags.forceVerification {
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

// readConfig attempts to interpret the bytes from an io.Reader as one of three possible
// structs: a build.Config, build.Result, or build.VerificationResult. All three of
// these structs contain complete build configuration needed to run a build.
//
// This is intended to make the system flexible: given any of these three things, you
// can attempt to reproduce the build they represent.
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
		b.buildResult = vr.Primary
		b.buildConfig = &vr.Primary.Config
		return vr.Primary.Config, nil
	}
	return build.Config{}, fmt.Errorf("not a build config, build result, or verification result")
}

// tryUnmarshalJSON attempts to intepret data as a T, and returns a T and true if successful,
// and returns a zero T and false otherwise.
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

// ensureAbs takes a path and ensures it's absolute relative to the current working directory.
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
