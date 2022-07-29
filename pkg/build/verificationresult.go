package build

import (
	"errors"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

// VerificationResult captures the result of a primary
// and local verification build together with easy-access
// hashes and an overall "reproduced correctly" boolean.
type VerificationResult struct {
	Primary             *Result
	Verification        *Result
	Hashes              crt.FileSetHashes
	ErrorMessage        string `json:",omitempty"`
	ReproducedCorrectly bool
}

func (vr *VerificationResult) Error() error {
	if vr.ErrorMessage == "" {
		return nil
	}
	return errors.New(vr.ErrorMessage)
}

// NewVerificationResult constructs a new VerificationResult ready for
// serialisation.
func NewVerificationResult(primary, verification Result) (*VerificationResult, error) {
	hashes, err := GetAllHashes(primary.Config, verification.Config)
	if err != nil {
		return nil, err
	}

	var errMessage string
	var reproduced bool
	hashErr := hashes.Error()
	if hashErr != nil {
		errMessage = hashErr.Error()
	} else {
		reproduced = true
	}

	return &VerificationResult{
		Primary:             &primary,
		Verification:        &verification,
		Hashes:              hashes,
		ErrorMessage:        errMessage,
		ReproducedCorrectly: reproduced,
	}, nil
}
