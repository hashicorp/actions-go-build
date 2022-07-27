package build

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

// Recorder is responsible for executing and logging build steps and
// constructing the build Result.
type Recorder struct {
	steps  []step
	result Result
	// nowFunc is usually time.Now but can be overridden
	// in tests.
	nowFunc func() time.Time
}

func NewRecorder(b Build) *Recorder {
	return &Recorder{
		result: Result{
			Config: b.Config(),
			Env:    b.Env(),
		},
		nowFunc: time.Now,
	}
}

func (br *Recorder) AddStep(desc string, action func() error) {
	br.steps = append(br.steps, step{desc, action})
}

func (br *Recorder) Run() Result {
	br.start()
	for _, s := range br.steps {
		if br.recordStep(s.desc, s.action); br.Failed() {
			break
		}
	}
	return br.Result()
}

func (br *Recorder) isFinished() bool {
	return br.result.Meta.Finish != (time.Time{})
}

func (br *Recorder) finish() {
	if !br.isFinished() {
		br.result.Meta.Finish = br.nowFunc()
		br.result.Meta.Duration = br.result.Meta.Finish.Sub(br.result.Meta.Start).String()
		br.result.Successful = br.result.err == nil
	}
}

func (br *Recorder) Result() Result {
	br.finish()
	return br.result
}

func (br *Recorder) Failed() bool {
	return br.result.err != nil
}

func (br *Recorder) RecordBin(path string) error {
	var err error
	br.result.Executable, err = getFileDetails(path)
	return err
}

func (br *Recorder) RecordZip(path string) error {
	var err error
	br.result.Zip, err = getFileDetails(path)
	return err
}

func (br *Recorder) start() *Recorder {
	br.result.Meta.Start = br.nowFunc()
	return br
}

func (br *Recorder) recordStep(desc string, step func() error) error {
	err := step()
	if err == nil {
		log.Printf("SUCCESS: %s", desc)
		return nil
	}
	// Add the step description to the error.
	err = fmt.Errorf("%s failed: %w", desc, err)
	br.result.err = err
	log.Printf("ERROR: %s", err)
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
