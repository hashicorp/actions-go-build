package build

import (
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

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
	Config       Config
	Env          []string
	Meta         Meta
	Zip          crt.File
	Executable   crt.File
	err          error
	ErrorMessage string `json:",omitempty"`
	Successful   bool
}

func (br Result) Error() error {
	return br.err
}

// Meta captures after-the-fact information about the build.
// This will be different between primary and verification builds.
type Meta struct {
	Start, Finish time.Time
	Duration      string
}

type step struct {
	desc   string
	action func() error
}
