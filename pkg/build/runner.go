package build

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

// Runner is responsible for executing and logging build steps and
// constructing the build Result.
type Runner struct {
	Settings
	build  Build
	result Result
	// nowFunc is usually time.Now but can be overridden
	// in tests.
	nowFunc func() time.Time
}

func NewRunner(b Build, opts ...Option) (*Runner, error) {
	s, err := newSettings(opts)
	if err != nil {
		return nil, err
	}
	return &Runner{
		Settings: s,
		build:    b,
		result:   Result{},
		nowFunc:  time.Now,
	}, nil
}

type StepFunc func() error

type Step struct {
	desc   string
	action StepFunc
}

func (br *Runner) Run() Result {
	br.Log("Beginning build, rooted at %q", br.build.Config().Paths.WorkDir)
	br.start()
	for _, s := range br.build.Steps() {
		if br.recordStep(s.desc, s.action); br.Failed() {
			break
		}
	}
	if !br.Failed() {
		br.recordStep("recording executable file details", func() error {
			return br.RecordBin(br.build.Config().Paths.BinPath)
		})
		br.recordStep("recording zip file details", func() error {
			return br.RecordZip(br.build.Config().Paths.ZipPath)
		})
	}
	return br.Result()
}

func (br *Runner) isFinished() bool {
	return br.result.Meta.Finish != (time.Time{})
}

func (br *Runner) finish() {
	if !br.isFinished() {
		br.result.Config = br.build.Config()
		br.result.Env = br.build.Env()
		br.result.Meta.Finish = br.nowFunc()
		br.result.Meta.Duration = br.result.Meta.Finish.Sub(br.result.Meta.Start).String()
		br.result.Successful = br.result.err == nil
	}
}

func (br *Runner) Result() Result {
	br.finish()
	return br.result
}

func (br *Runner) Failed() bool {
	return br.result.err != nil
}

func (br *Runner) RecordBin(path string) error {
	var err error
	br.result.Executable, err = getFileDetails(path)
	return err
}

func (br *Runner) RecordZip(path string) error {
	var err error
	br.result.Zip, err = getFileDetails(path)
	return err
}

func (br *Runner) start() *Runner {
	br.result.Meta.Start = br.nowFunc()
	return br
}

func (br *Runner) recordStep(desc string, step func() error) error {
	br.Debug("%s: starting", desc)
	err := step()
	if err == nil {
		br.Log("%s: ok", desc)
		return nil
	}
	// Add the step description to the error.
	err = fmt.Errorf("%s: %w", desc, err)
	br.result.err = err
	br.result.ErrorMessage = err.Error()
	br.Log("%s: failed: %s", desc, err)
	return err
}

func getFileDetails(path string) (crt.File, error) {
	f := crt.File{
		Name:         filepath.Base(path),
		OriginalPath: path,
	}
	fi, err := os.Stat(path)
	if err != nil {
		return f, err
	}
	f.Size = fi.Size()
	f.SHA256Sum, err = digest.FileSHA256Hex(path)
	return f, err
}
