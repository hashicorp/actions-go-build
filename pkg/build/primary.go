package build

// Primary is the primary build. This is run in the current working
// directory, using whatever files are present.
type Primary struct {
	core
}

func NewPrimary(cfg Config, opts ...Option) (Build, error) {
	return New(cfg, opts...)
}
