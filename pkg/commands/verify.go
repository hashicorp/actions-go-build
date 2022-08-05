package commands

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/artdarek/go-unzip"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

type verifyOpts struct {
	present    presenter
	resultFile string
	build      buildFlags
}

func (opts *verifyOpts) ReadEnv() error { return cli.ReadEnvAll(&opts.present) }

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &opts.present, &opts.build)
	fs.StringVar(&opts.resultFile, "result", "", "path to the json build result file to verify")
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
		return fmt.Errorf("result is dirty (source hash != revision); we can't verify dirty builds")
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
	response, err := http.Get(sourceURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if _, err := io.Copy(destFile, response.Body); err != nil {
		return err
	}
	if err := destFile.Close(); err != nil {
		return err
	}

	// Extract the downloaded zip file.
	sourcePath := filepath.Join(tmpDir, br.Config.Product.Name)
	if err := fs.Mkdir(sourcePath); err != nil {
		return err
	}
	if err := unzip.New(destFilePath, sourcePath).Extract(); err != nil {
		return err
	}

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
