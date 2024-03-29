// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/actions-go-build/internal/unzipper"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
)

// RemoteBuild is a build where the source code is hosted remotely.
// This can be a "primary" or verification build.
type RemoteBuild struct {
	*core
	sourceURL string
	cacheID   string
}

func NewRemoteBuild(c Config, options ...Option) (Build, error) {
	if c.Product.IsDirty() {
		return nil, fmt.Errorf("cannot verify a dirty build remotely")
	}

	parts := strings.Split(c.Product.Module, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("module %q not supported, must be in the format %q", c.Product.Module, "github.com/<user>/<repo>...")
	}
	user, repo, revision := parts[1], parts[2], c.Product.Revision

	sourceURL := fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", user, repo, revision)
	core, err := newCore("remote build", c, options...)
	if err != nil {
		return nil, err
	}
	if err := core.UpdateBuildRoot(); err != nil {
		return nil, err
	}
	return &RemoteBuild{
		core:      core,
		sourceURL: sourceURL,
	}, nil
}

func (rb *RemoteBuild) Steps() []Step {

	var sourceDLDir, sourceArchivePath string

	pre := []Step{
		newStep("change build root to temporary directory", func() error {
			if rb.IsVerification() {
				return rb.ChangeToVerificationRoot()
			}
			return rb.ChangeToPrimaryRoot()
		}),
		newStep("create temporary paths", func() error {
			sourceDLDir = rb.Dirs().SourceDownloadDir()
			return fs.MkdirEmpty(sourceDLDir)
		}),
		newStep(fmt.Sprintf("get %s", rb.sourceURL), func() error {
			c := rb.Config()
			sourceArchiveName := fmt.Sprintf("%s-%s.zip", c.Product.Name, c.Product.Revision)
			sourceArchivePath = filepath.Join(sourceDLDir, sourceArchiveName)
			destFile, err := os.Create(sourceArchivePath)
			if err != nil {
				return err
			}
			var closeErr error
			defer func() { closeErr = destFile.Close() }()
			response, err := http.Get(rb.sourceURL)
			if err != nil {
				return err
			}
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("%s", response.Status)
			}
			var bodyCloseErr error
			defer func() { bodyCloseErr = response.Body.Close() }()
			if _, err := io.Copy(destFile, response.Body); err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
			return bodyCloseErr
		}),
		newStep("extract source code to temporary directory", func() error {
			// Extract the downloaded zip file directly in the same dir as the zip.
			// These zips contain a directory that contains all the code, so we'll
			// use that directory as the build root.
			return unzipper.New(rb.Debug).Unzip(sourceArchivePath, sourceDLDir)
		}),
		newStep("move source code to build root", func() error {
			c := rb.Config()
			// This innerDirName is GitHub-specific.
			innerDirName := fmt.Sprintf("%s-%s", path.Base(c.Product.Repository), c.Product.Revision)
			sourcePath := filepath.Join(sourceDLDir, innerDirName)
			return fs.Move(sourcePath, rb.Config().Paths.WorkDir)
		}),
	}

	return append(pre, rb.core.Steps()...)
}
