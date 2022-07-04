package config

// BuildConfig contains the
type BuildConfig struct {
	// Product is the logical product being built.
	Product Product
	// ProductVersionMeta is the metadata component of the product version.
	// E.g. "ent" or "ent.fips".
	ProductVersionMeta string
	// WorkDir is the absolute directory to run the build instructions in.
	WorkDir string
	// TargetDir is the absolute path to the dir where any other files
	// needed to be included in the zip file should be placed.
	TargetDir string
	// BinPath is the path to the executable binary the instructions must create.
	BinPath string
	// ZipPath is the path to the zip file that will be created.
	ZipPath string
	// Instructions is the build instructions.
	Instructions string
	TargetOS     string
	TargetArch   string

	// TODO: Consider removing these fields if possible, we should be able
	// to derive them from other context.
	ZipDir  string
	MetaDir string
}
