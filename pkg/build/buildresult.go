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

// DoubleBuildResult captures the result of a primary
// and local verification build together.
type DoubleBuildResult struct {
	Primary      *Result
	Verification *Result
	Hashes       crt.FileSetHashes
}

func NewDoubleBuildResult(primary, verification Result) (*DoubleBuildResult, error) {
	hashes, err := GetAllHashes(primary.Config, verification.Config)
	if err != nil {
		return nil, err
	}

	return &DoubleBuildResult{
		Primary:      &primary,
		Verification: &verification,
		Hashes:       hashes,
	}, nil
}

// Inputs represents the fixed inuputs to the build.
// These are identical for both the primary and verification
// build.
type Inputs struct {
	Product    crt.Product
	Parameters Parameters
}

// Result captures a single binary build. It's used for
// both primary and verification builds.
// Note that the Config will be different for each of
// them because it contains build-host-specific paths.
type Result struct {
	Config       BuildConfig
	Env          []string
	Meta         Meta
	Zip          crt.File
	Executable   crt.File
	err          error
	ErrorMessage string
	Successful   bool
}

func (br Result) Error() error {
	return br.err
}

// Meta captures after-the-fact information about the build.
// This will be different between primary and verification builds.
type Meta struct {
	Start, Finish time.Time
}

type step struct {
	desc   string
	action func() error
}

type Recorder struct {
	steps  []step
	result Result
	// nowFunc is usually time.Now but can be overridden
	// in tests.
	nowFunc func() time.Time
}

func NewRecorder(c BuildConfig) *Recorder {
	return &Recorder{
		result: Result{
			Config: c,
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

func (br *Recorder) Result() Result {
	br.result.Meta.Finish = br.nowFunc()
	br.result.Successful = br.result.err == nil
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
