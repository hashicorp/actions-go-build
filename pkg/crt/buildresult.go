package crt

// BuildResult captures the result of a local primary and
// verification build together.
type BuildResult struct {
	Inputs       BuildInputs
	Primary      *Build
	Verification *Build
}

type BuildInputs struct {
	Product    Product
	Parameters BuildParameters
}

// Build captures a single binary build. It's used for
// both primary and verification builds.
type Build struct {
	Config     BuildConfig
	Zip        File
	Executable File
}

// File is a file produced during the build.
type File struct {
	// Name is the base name of the file.
	Name string
	// OriginalPath is the absolute path this file was written to initially.
	OriginalPath string
	// Size is the size of the file in bytes.
	Size int64
	// SHA256Sum is the digest of the file.
	SHA256Sum string
	// URL is the URL of the uploaded artifact.
	// This may be empty if the file has not been uploaded
	// as an artifact.
	URL string
}
