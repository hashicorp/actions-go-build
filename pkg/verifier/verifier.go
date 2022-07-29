package verifier

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type ResultSource interface {
	Result() (build.Result, error)
}

type Verifier struct {
	primary, verification ResultSource
}

func New(primary, verification ResultSource) *Verifier {
	return &Verifier{
		primary:      primary,
		verification: verification,
	}
}

func loadResult(name string, rs ResultSource) (*build.Result, error) {
	r, err := rs.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get %s build result: %w", name, err)
	}
	if err := r.Error(); err != nil {
		return nil, fmt.Errorf("%s build failed: %w", name, r.Error())
	}
	return &r, nil
}

// Verify returns a VerificationResult which may or may not be affirmative.
// It returns an error when issues occur discovering that result, not
// when the result itself says that the reproduction didn't work.
// You still need to query the result to find out if it was successful.
func (v *Verifier) Verify() (*build.VerificationResult, error) {
	pr, err := loadResult("primary", v.primary)
	if err != nil {
		return nil, err
	}
	vr, err := loadResult("verification", v.verification)
	if err != nil {
		return nil, err
	}
	return v.makeResult(pr, vr)
}

func (v *Verifier) makeResult(pr, vr *build.Result) (*build.VerificationResult, error) {
	// Exit early if we're comparing apples with oranges.
	if diff := cmp.Diff(pr.Config.Product, vr.Config.Product); diff != "" {
		return nil, fmt.Errorf("product details are not identical: %s", diff)
	}
	if diff := cmp.Diff(pr.Config.Parameters, vr.Config.Parameters); diff != "" {
		return nil, fmt.Errorf("build parameters are not identical: %s", diff)
	}

	binHashes, binErr := fileHashes("executable", pr.Executable, vr.Executable)
	zipHashes, zipErr := fileHashes("zip", pr.Zip, vr.Zip)

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

	return &build.VerificationResult{
		Primary:             pr,
		Verification:        vr,
		Hashes:              hashes,
		ErrorMessage:        errMessage,
		ReproducedCorrectly: err == nil,
	}, nil
}

func fileHashes(desc string, p, v crt.File) (crt.FileHashes, error) {
	match := p.SHA256Sum == v.SHA256Sum
	var err error
	if p.Name != v.Name {
		err = fmt.Errorf("names are different: %q and %q", p.Name, v.Name)
	} else if p.Size != v.Size {
		err = fmt.Errorf("sizes are different: %v and %v", p.Size, v.Size)
	} else if !match {
		err = fmt.Errorf("digests are different: %q and %q", p.SHA256Sum, v.SHA256Sum)
	}
	return crt.FileHashes{
		Name:        p.Name,
		Description: desc,
		SHA256: crt.HashPair{
			Primary:      p.SHA256Sum,
			Verification: v.SHA256Sum,
			Match:        match,
		},
	}, err
}
