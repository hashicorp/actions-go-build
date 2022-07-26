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
)

type ResultWriter struct {
	github GitHubOpts
	file   *os.File
}

func (brw *ResultWriter) ReadEnv() error         { return cli.ReadEnvAll(&brw.github) }
func (brw *ResultWriter) Flags(fs *flag.FlagSet) { cli.FlagsAll(fs, &brw.github) }

func buildResultFilename(br *build.Result) string {
	return fmt.Sprintf("%s.buildresult.json", filepath.Base(br.Config.Paths.ZipPath))
}

func (brw *ResultWriter) buildResultWriter(br *build.Result) (io.Writer, string, error) {
	var filename string
	if !brw.github.GitHubMode {
		return os.Stdout, filename, nil
	}
	filename = buildResultFilename(br)
	w, err := brw.multiWriter(filename)
	return w, filename, err
}

// WriteBuildResult returns the path written.
func (brw *ResultWriter) WriteBuildResult(br *build.Result) (string, error) {
	writer, filename, err := brw.buildResultWriter(br)
	if err != nil {
		return filename, err
	}
	var closeErr error
	defer func() { closeErr = brw.Close() }()
	if err := writeJSON(writer, br); err != nil {
		return filename, err
	}
	return filename, closeErr
}

func doubleBuildResultFilename(br *build.DoubleBuildResult) string {
	return fmt.Sprintf("%s.doubleresult.json", filepath.Base(br.Primary.Config.Paths.ZipPath))
}

func (brw *ResultWriter) doubleBuildResultWriter(br *build.DoubleBuildResult) (io.Writer, string, error) {
	var filename string
	if !brw.github.GitHubMode {
		return os.Stdout, filename, nil
	}
	filename = doubleBuildResultFilename(br)
	w, err := brw.multiWriter(filename)
	return w, filename, err
}

// WriteDoubleBuildResult returns the path written.
func (brw *ResultWriter) WriteDoubleBuildResult(br *build.DoubleBuildResult) (string, error) {
	writer, filename, err := brw.doubleBuildResultWriter(br)
	if err != nil {
		return filename, err
	}
	var closeErr error
	defer func() { closeErr = brw.Close() }()
	if err := writeJSON(writer, br); err != nil {
		return filename, err
	}
	return filename, closeErr
}

func writeJSON(w io.Writer, thing any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(thing)
}

func (brw *ResultWriter) multiWriter(filename string) (io.Writer, error) {
	var err error
	if brw.file, err = os.Create(filename); err != nil {
		return nil, err
	}
	return io.MultiWriter(os.Stdout, brw.file), nil
}

func (brw *ResultWriter) Close() error {
	if brw.file == nil {
		return nil
	}
	return brw.file.Close()
}
