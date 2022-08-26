package build

// Primary is the primary build. This is run in the current working
// directory, using whatever files are present.
type Primary struct {
	*core
}

func NewPrimary(cfg Config, opts ...Option) (Build, error) {
	opts = append(opts, AsPrimaryBuild())
	core, err := newCore("primary", cfg, opts...)
	if err != nil {
		return nil, err
	}
	return &Primary{core: core}, nil
}

func (p *Primary) Kind() string { return "primary" }
