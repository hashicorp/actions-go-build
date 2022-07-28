package opts

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
)

type ResultWriter struct {
	github   GitHubOpts
	filename string
	file     *os.File
}

func (brw *ResultWriter) ReadEnv() error {
	return cli.ReadEnvAll(&brw.github)
}

func (brw *ResultWriter) Flags(fs *flag.FlagSet) {
	cli.FlagsAll(fs, &brw.github)
	fs.StringVar(&brw.filename, "output", "", "overwrite file path to write JSON results")
}

// WriteBuildResult returns the path written.
func (brw *ResultWriter) WriteBuildResult(br build.Result) (string, error) {
	return writeResult(brw, br, buildResultFilename)
}

func (brw *ResultWriter) WriteVerificationResult(vr *build.VerificationResult) (string, error) {
	return writeResult(brw, vr, verificationResultFilename)
}
func (brw *ResultWriter) closeFile() error {
	if brw.file == nil {
		return nil
	}
	return brw.file.Close()
}

func buildResultFilename(br build.Result) string {
	return fmt.Sprintf("%s.buildresult.json", filepath.Base(br.Config.Paths.ZipPath))
}

func (brw *ResultWriter) makeWriter(defaultFilename string) (io.Writer, string, error) {
	filename := brw.filename
	if filename == "" {
		filename = defaultFilename
	}
	if err := fs.Mkdir(filepath.Dir(filename)); err != nil {
		return nil, "", err
	}
	var err error
	if brw.file, err = os.Create(filename); err != nil {
		return nil, "", err
	}
	if !brw.github.GitHubMode {
		return brw.file, filename, nil
	}
	return io.MultiWriter(os.Stdout, brw.file), filename, nil
}

func verificationResultFilename(br *build.VerificationResult) string {
	return fmt.Sprintf("meta/%s.local-verification-result.json", filepath.Base(br.Primary.Config.Paths.ZipPath))
}

func writeResult[T any](brw *ResultWriter, a T, nameFunc func(T) string) (string, error) {
	writer, filename, err := brw.makeWriter(nameFunc(a))
	if err != nil {
		return filename, err
	}
	var closeErr error
	defer func() { closeErr = brw.closeFile() }()
	if err := writeJSON(writer, a); err != nil {
		return filename, err
	}
	return filename, closeErr
}

func writeJSON(w io.Writer, thing any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(thing)
}
