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
	Dirty               bool
	ReproducedCorrectly bool
}

func (vr *VerificationResult) Error() error {
	if vr.ErrorMessage == "" {
		return nil
	}
	return errors.New(vr.ErrorMessage)
}

func (vr *VerificationResult) IsFromCache() bool { return false }
