package crt

// File is a file produced during the build.
type File struct {
	// Name is the base name of the file.
	Name string
	// OriginalPath is the absolute path this file was written to.
	OriginalPath string
	// Size is the size of the file in bytes.
	Size int64
	// SHA256Sum is the digest of the file.
	SHA256Sum string
}
