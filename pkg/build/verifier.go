package build

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type ResultSource interface {
	Result() (Result, error)
}

type Verifier struct {
	Settings
	primary, verification ResultSource
}

func NewVerifier(primary, verification ResultSource, opts ...Option) (*Verifier, error) {
	s, err := newSettings(opts)
	if err != nil {
		return nil, err
	}
	v := &Verifier{
		Settings:     s,
		primary:      primary,
		verification: verification,
	}
	v.Debug("initialised")
	return v, nil
}

// Verify returns a VerificationResult which may or may not be affirmative.
// It returns an error when issues occur discovering that result, not
// when the result itself says that the reproduction didn't work.
// You still need to query the result to find out if it was successful.
func (v *Verifier) Verify() (*VerificationResult, error) {
	v.Debug("beginning verification")
	pr, err := v.loadResult("primary", v.primary)
	if err != nil {
		return nil, err
	}
	v.Debug("got primary build result")

	if pr.Config.Product.IsDirty() {
		v.Loud("WARNING: Primary result is dirty: source hash (%s...) != revision (%s...)",
			pr.Config.Product.SourceHash[:8], pr.Config.Product.Revision[:8])
	}

	vr, err := v.loadResult("verification", v.verification)
	if err != nil {
		return nil, err
	}
	v.Debug("got verification build result")
	return v.verificationResult(pr, vr)
}

func (v *Verifier) loadResult(name string, rs ResultSource) (*Result, error) {
	v.Debug("Getting %s build result", name)
	r, err := rs.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get %s build result: %w", name, err)
	}
	if err := r.Error(); err != nil {
		return nil, fmt.Errorf("%s build failed: %w", name, r.Error())
	}
	return &r, nil
}

func (v *Verifier) verificationResult(pr, vr *Result) (*VerificationResult, error) {
	v.Debug("Returning verification result.")
	// Exit early if we're comparing apples with oranges.
	if diff := cmp.Diff(pr.Config.Product, vr.Config.Product); diff != "" {
		return nil, fmt.Errorf("product details are not identical: %s", diff)
	}
	if diff := cmp.Diff(pr.Config.Parameters, vr.Config.Parameters); diff != "" {
		return nil, fmt.Errorf("build parameters are not identical: %s", diff)
	}

	binHashes, binErr := v.fileHashes("executable", pr.Executable, vr.Executable)
	zipHashes, zipErr := v.fileHashes("zip", pr.Zip, vr.Zip)

	var err error
	if binErr != nil {
		err = binErr
	} else if zipErr != nil {
		err = zipErr
	}
	var errMessage string
	if err != nil {
		errMessage = err.Error()
	}

	hashes := crt.NewFileSetHashes(binHashes, zipHashes)

	return &VerificationResult{
		Primary:             pr,
		Verification:        vr,
		Hashes:              hashes,
		ErrorMessage:        errMessage,
		ReproducedCorrectly: err == nil,
	}, nil
}

func (v *Verifier) fileHashes(desc string, pf, vf crt.File) (crt.FileHashes, error) {
	v.Debug("Comparing primary and verification versions of %s file: %s", desc, pf.Name)
	match := pf.SHA256Sum == vf.SHA256Sum
	var err error
	if pf.Name != vf.Name {
		err = fmt.Errorf("names are different: %q and %q", pf.Name, vf.Name)
	} else if pf.Size != vf.Size {
		err = fmt.Errorf("sizes are different: %v and %v", pf.Size, vf.Size)
	} else if !match {
		err = fmt.Errorf("digests are different: %q and %q", pf.SHA256Sum, vf.SHA256Sum)
	}
	return crt.FileHashes{
		Name:        pf.Name,
		Description: desc,
		SHA256: crt.HashPair{
			Primary:      pf.SHA256Sum,
			Verification: vf.SHA256Sum,
			Match:        match,
		},
	}, err
}
