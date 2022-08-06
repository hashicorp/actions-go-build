package build

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/artdarek/go-unzip"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
)

// RemoteVerification is a build where the source code is hosted remotely.
// This is the kind of build we run when verifying a build result is reproducible.
type RemoteVerification struct {
	*core
	sourceURL string
	cacheID   string
}

func NewRemoteVerification(sourceURL string, cfg Config, options ...Option) (Build, error) {
	core, err := newCore("remote-verification-build", cfg, options...)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}
	if u.Host != "github.com" {
		// We only support GitHub for now because the logic for extracting GitHub flavoured
		// source zips is baked in here. We could add handling for other sources pretty trivially.
		return nil, fmt.Errorf("currently only source code hosted on GitHub is supported")
	}
	return &RemoteVerification{
		core:      core,
		sourceURL: sourceURL,
	}, nil
}

func (lv *RemoteVerification) Kind() string { return "local verification" }

func (lv *RemoteVerification) Steps() []Step {

	c := lv.core.Config()

	tmpDir := filepath.Join(os.TempDir(), "actions-go-build", "remote-verification", c.Product.Name, c.Product.Revision)
	fileName := fmt.Sprintf("%s-%s.zip", c.Product.Name, c.Product.Revision)
	destFilePath := filepath.Join(tmpDir, fileName)

	// This innerDirName is GitHub-specific.
	innerDirName := fmt.Sprintf("%s-%s", path.Base(c.Product.Repository), c.Product.Revision)
	sourcePath := filepath.Join(tmpDir, innerDirName)

	pre := []Step{
		newStep("creating temporary directory to run build in", func() error {
			exists, err := fs.DirExists(tmpDir)
			if err != nil {
				return err
			}
			if exists {
				if err := os.RemoveAll(tmpDir); err != nil {
					return err
				}
			}
			return fs.Mkdir(tmpDir)
		}),
		newStep(fmt.Sprintf("downloading source code from %s", lv.sourceURL), func() error {
			destFile, err := os.Create(destFilePath)
			if err != nil {
				return err
			}
			var closeErr error
			defer func() { closeErr = destFile.Close() }()
			response, err := http.Get(lv.sourceURL)
			if err != nil {
				return err
			}
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("unable to download source code: %s", response.Status)
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
		newStep("extracting source code to temporary directory", func() error {
			// Extract the downloaded zip file directly in the same dir as the zip.
			// These zips contain a directory that contains all the code, so we'll
			// use that directory as the build root.
			return unzip.New(destFilePath, tmpDir).Extract()
		}),
		newStep("changing build root to temporary directory", func() error {
			// Change our build root to the downloaded source dir.
			return lv.core.ChangeRoot(sourcePath)
		}),
	}

	return append(pre, lv.core.Steps()...)
}
