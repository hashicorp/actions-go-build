package build

import (
	"fmt"
	"io"
	"net/http"
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

func NewRemoteVerification(c Config, options ...Option) (Build, error) {

	sourceURL := fmt.Sprintf("https://github.com/%s/archive/%s.zip", c.Product.Repository, c.Product.Revision)

	core, err := newCore("remote-verification-build", c, options...)
	if err != nil {
		return nil, err
	}

	if err := core.ChangeToVerificationRoot(); err != nil {
		return nil, err
	}

	return &RemoteVerification{
		core:      core,
		sourceURL: sourceURL,
	}, nil
}

func (lv *RemoteVerification) Kind() string { return "verification" }

func (lv *RemoteVerification) Steps() []Step {

	var sourceDLDir, sourceArchivePath string

	pre := []Step{
		newStep("change build root to temporary directory", func() error {
			return lv.ChangeToVerificationRoot()
		}),
		newStep("create temporary paths", func() error {
			c := lv.Config()
			sourceDLDir = filepath.Join(os.TempDir(), "actions-go-build", "rv", c.Product.Name, c.Product.Revision)
			return fs.MkdirEmpty(sourceDLDir)
		}),
		newStep(fmt.Sprintf("get %s", lv.sourceURL), func() error {
			c := lv.Config()
			sourceArchiveName := fmt.Sprintf("%s-%s.zip", c.Product.Name, c.Product.Revision)
			sourceArchivePath = filepath.Join(sourceDLDir, sourceArchiveName)
			destFile, err := os.Create(sourceArchivePath)
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
			return unzip.New(sourceArchivePath, sourceDLDir).Extract()
		}),
		newStep("move source code to verification root", func() error {
			c := lv.Config()
			// This innerDirName is GitHub-specific.
			innerDirName := fmt.Sprintf("%s-%s", path.Base(c.Product.Repository), c.Product.Revision)
			sourcePath := filepath.Join(sourceDLDir, innerDirName)
			return fs.Move(sourcePath, lv.Config().Paths.WorkDir)
		}),
	}

	return append(pre, lv.core.Steps()...)
}
