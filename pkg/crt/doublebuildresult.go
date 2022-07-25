package crt

// DoubleBuildResult captures the result of a primary
// and local verification build together.
type DoubleBuildResult struct {
	Primary      *BuildResult
	Verification *BuildResult
	Hashes       FileSetHashes
}

func NewDoubleBuildResult(primary, verification BuildResult) (*DoubleBuildResult, error) {
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
