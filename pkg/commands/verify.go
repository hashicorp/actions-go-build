package commands

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/artdarek/go-unzip"
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type verifyOpts struct {
	present    presenter
	build      buildFlags
	resultFile string
}

func (opts *verifyOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present) }

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.present, &opts.build)
}

func (opts *verifyOpts) ParseArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("argument missing: path to json build result file")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: exactly one required")
	}
	opts.resultFile = args[0]
	return nil
}

var Verify = cli.LeafCommand("verify", "verify a build result", func(opts *verifyOpts) error {

	if opts.resultFile == "" {
		return fmt.Errorf("verify requires the -result flag to be set")
	}

	br, err := json.ReadFile[build.Result](opts.resultFile)
	if err != nil {
		return err
	}

	if br.Config.Product.IsDirty() {
		log.Info("WARNING: Result is dirty: source hash != revision")
	}

	// Update the build paths to a temp dir to run the verification build in.
	tmpDir, err := os.MkdirTemp("", "verification-build.*")
	if err != nil {
		return err
	}

	// Download the source code to be built.
	sourceURL := fmt.Sprintf("https://github.com/%s/archive/%s.zip", br.Config.Product.Repository, br.Config.Product.Revision)
	fileName := fmt.Sprintf("%s-%s.zip", br.Config.Product.Name, br.Config.Product.Revision)
	destFilePath := filepath.Join(tmpDir, fileName)
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	log.Info("Downloading %s", sourceURL)
	response, err := http.Get(sourceURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if _, err := io.Copy(destFile, response.Body); err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to download source code: %s", response.Status)
	}
	if err := destFile.Close(); err != nil {
		return err
	}

	// Extract the downloaded zip file directly in the same dir as the zip.
	// These zips contain a directory that contains all the code, so we'll
	// use that directory as the build root.
	if err := unzip.New(destFilePath, tmpDir).Extract(); err != nil {
		return err
	}

	innerDirName := fmt.Sprintf("%s-%s", path.Base(br.Config.Product.Repository), br.Config.Product.Revision)
	sourcePath := filepath.Join(tmpDir, innerDirName)

	// Change our build root to the downloaded source dir.
	c, err := br.Config.ChangeRoot(sourcePath)
	if err != nil {
		return err
	}

	b, err := build.New(c)
	if err != nil {
		return err
	}
	m := opts.build.newManager(b)

	verifier := verifier.New(br, m)

	result, err := verifier.Verify()
	if err != nil {
		return err
	}

	return opts.present.result("Verification result", result)

}).WithHelp(`
Verify that a build result is reproducible.

This command accepts a build result JSON file, uses it to run a new verification
build, and compares the results.
`)
