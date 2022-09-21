package build

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/go-version"
)

// Parameters are the set of build inputs that should be enough (along with
// the Product) to reproduce a build. Changes to these should result in different
// build outputs.
type Parameters struct {
	// GoVersion is the version of the Go toolchain to run thid build with.
	GoVersion string `env:"GO_VERSION"`
	// Instructions are the build instructions (a bash script).
	Instructions string `env:"INSTRUCTIONS"`
	// OS is the target OS for this build.
	OS string `env:"OS"`
	// Arch is the target Architecture for this build.
	Arch string `env:"ARCH"`
	// ZipName is the name of the zip file to create.
	ZipName string `env:"ZIP_NAME"`
}

func (bp Parameters) Init(p crt.Product) (Parameters, error) {
	return bp.trimSpace().setDefaults(p)
}

func (bp Parameters) trimSpace() Parameters {
	trim(&bp.GoVersion, &bp.Instructions, &bp.OS, &bp.Arch, &bp.ZipName)
	return bp
}

func (bp Parameters) setDefaults(p crt.Product) (Parameters, error) {
	if bp.GoVersion == "" {
		bp.GoVersion = strings.TrimPrefix(runtime.Version(), "go")
	}
	if bp.OS == "" {
		bp.OS = runtime.GOOS
	}
	if bp.Arch == "" {
		bp.Arch = runtime.GOARCH
	}
	if bp.ZipName == "" {
		bp.ZipName = bp.defaultZipName(p)
	}
	if bp.Instructions == "" {
		var err error
		bp.Instructions, err = bp.defaultInstructions(p)
		if err != nil {
			return bp, err
		}
	}
	return bp, nil
}

func (bp Parameters) defaultZipName(p crt.Product) string {
	return fmt.Sprintf("%s_%s_%s_%s.zip", p.Name, p.Version.Full, bp.OS, bp.Arch)
}

func (bp Parameters) defaultInstructions(p crt.Product) (string, error) {
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
		// away from the repository. It's currently needed because it breaks
		// reproducibility when buildvcs=true.
		flags = append(flags, "-buildvcs=false")
	}
	if p.Module != "" {
		ldFlags := fmt.Sprintf(`"%s"`, defaultLDFlags(p))
		flags = append(flags, "-ldflags", ldFlags)
	}
	return strings.Join(flags, " "), nil
}

type kv struct {
	k, v string
}

func defaultLDFlags(p crt.Product) string {
	var ldflags []string
	for _, kv := range []kv{
		{"Repository", p.Repository},
		{"Module", p.Module},
		{"Name", p.Name},
		{"CoreName", p.CoreName},
		{"ExecutableName", p.ExecutableName},
		{"VersionFull", p.Version.Full},
		{"VersionCore", p.Version.Core},
		{"VersionMeta", p.Version.Meta},
		{"Revision", p.Revision},
		{"RevisionTime", p.RevisionTime},
		{"SourceHash", p.SourceHash},
	} {
		ldflags = append(ldflags, mkLDFlag(p, kv.k, kv.v))
	}
	return strings.Join(ldflags, " ")
}

func mkLDFlag(p crt.Product, name, value string) string {
	return fmt.Sprintf("-X '%s/product.%s=%s'", p.Module, name, value)
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
