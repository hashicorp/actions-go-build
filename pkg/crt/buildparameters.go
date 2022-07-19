package crt

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
)

// BuildParameters are the set of build inputs that should be enough (along with
// the Product) to reproduce a build. Changes to these should result in different
// build outputs.
type BuildParameters struct {
	// GoVersion is the version of the Go toolchain to run thid build with.
	GoVersion string `env:"GO_VERSION"`
	// Instructions are the build instructions (a bash script).
	Instructions string `env:"INSTRUCTIONS"`
	// OS is the target OS for this build.
	OS string `env:"OS"`
	// Arch is the target Architecture for this build.
	Arch string `env:"ARCH"`
}

func (bp BuildParameters) Init(p Product) (BuildParameters, error) {
	return bp.trimSpace().setDefaults(p)
}

func (bp BuildParameters) trimSpace() BuildParameters {
	trim(&bp.GoVersion, &bp.Instructions, &bp.OS, &bp.Arch)
	return bp
}

func (bp BuildParameters) setDefaults(p Product) (BuildParameters, error) {
	if bp.GoVersion == "" {
		bp.GoVersion = strings.TrimPrefix(runtime.Version(), "go")
	}
	if bp.Instructions == "" {
		var err error
		bp.Instructions, err = bp.defaultInstructions(p)
		if err != nil {
			return bp, err
		}
	}
	if bp.OS == "" {
		bp.OS = runtime.GOOS
	}
	if bp.Arch == "" {
		bp.Arch = runtime.GOARCH
	}
	return bp, nil
}

func (bp BuildParameters) defaultInstructions(p Product) (string, error) {
	var flags []string
	flags = append(flags, "go", "build")
	flags = append(flags, "-o", `"$BIN_PATH"`)
	flags = append(flags, "-trimpath")
	gt, err := goVersion118OrGreater(bp.GoVersion)
	if err != nil {
		return "", err
	}
	if gt {
		// TODO: See if we can remove this by moving the output directories
		// away from the repository.
		flags = append(flags, "-buildvcs=false")
	}
	return strings.Join(flags, " "), nil
}

func goVersion118OrGreater(vs string) (bool, error) {
	goVersion, err := version.NewVersion(vs)
	if err != nil {
		return false, fmt.Errorf("parsing go version %q: %w", vs, err)
	}
	go118, err := version.NewVersion("1.18")
	if err != nil {
		return false, fmt.Errorf("parsing go version 1.18: %w", err)
	}
	return goVersion.GreaterThanOrEqual(go118), nil
}

func trim(ss ...*string) {
	for _, s := range ss {
		*s = strings.TrimSpace(*s)
	}
}
