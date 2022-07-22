package crt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/digest"
)

// DoubleBuildResult captures the result of a local primary and
// verification build together.
type DoubleBuildResult struct {
	Inputs       BuildInputs
	Primary      *BuildResult
	Verification *BuildResult
}

// BuildInputs represents the fixed inuputs to the build.
// These are identical for both the primary and verification
// build.
type BuildInputs struct {
	Product    Product
	Parameters BuildParameters
}

// BuildResult captures a single binary build. It's used for
// both primary and verification builds.
// Note that the Config will be different for each of
// them because it contains build-host-specific paths.
type BuildResult struct {
	Config       BuildConfig
	Meta         BuildMeta
	Zip          File
	Executable   File
	err          error
	ErrorMessage string
}

func (br BuildResult) Error() error {
	return br.err
}

// BuildMeta captures after-the-fact information about the build.
// This will be different between primary and verification builds.
type BuildMeta struct {
	Start, Finish time.Time
}

type step struct {
	desc   string
	action func() error
}

type BuildRecorder struct {
	steps  []step
	result BuildResult
	// nowFunc is usually time.Now but can be overridden
	// in tests.
	nowFunc func() time.Time
}

func NewBuildRecorder(c BuildConfig) *BuildRecorder {
	return &BuildRecorder{
		result: BuildResult{
			Config: c,
		},
		nowFunc: time.Now,
	}
}

func (br *BuildRecorder) Start() *BuildRecorder {
	br.result.Meta.Start = br.nowFunc()
	return br
}

func (br *BuildRecorder) RecordBin(path string) error {
	var err error
	br.result.Executable, err = getFileDetails(path)
	return err
}

func (br *BuildRecorder) RecordZip(path string) error {
	var err error
	br.result.Zip, err = getFileDetails(path)
	return err
}

func (br *BuildRecorder) RecordStep(desc string, step func() error) error {
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

func getFileDetails(path string) (File, error) {
	f := File{
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

func (br *BuildRecorder) AddStep(desc string, action func() error) {
	br.steps = append(br.steps, step{desc, action})
}

func (br *BuildRecorder) Run() BuildResult {
	for _, s := range br.steps {
		if br.RecordStep(s.desc, s.action); br.Failed() {
			break
		}
	}
	return br.Result()
}

func (br *BuildRecorder) Result() BuildResult {
	br.result.Meta.Finish = br.nowFunc()
	return br.result
}

func (br *BuildRecorder) Failed() bool {
	return br.result.err != nil
}
