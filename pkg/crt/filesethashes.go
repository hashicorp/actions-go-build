package crt

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/digest"
)

func GetAllHashes(primary, verification BuildConfig) (FileSetHashes, error) {
	getBinPath := func(bc BuildConfig) string { return bc.Paths.BinPath }
	getZipPath := func(bc BuildConfig) string { return bc.Paths.ZipPath }

	var fsh FileSetHashes
	var err error

	if fsh.Bin, err = getHashes(primary, verification, getBinPath); err != nil {
		return fsh, err
	}

	if fsh.Zip, err = getHashes(primary, verification, getZipPath); err != nil {
		return fsh, err
	}

	return fsh, nil

}

type FileHashes struct {
	Primary, Verification string
}

// mismatch returns true if the hashes are different, or if they are both empty.
func (fh FileHashes) mismatch() bool {
	return fh.Primary != fh.Verification && fh.Primary != ""
}

type FileSetHashes struct {
	Bin FileHashes
	Zip FileHashes
}

func (fsh FileSetHashes) Error() error {
	if fsh.Bin.mismatch() {
		return fmt.Errorf("executable file mismatch")
	}
	if fsh.Zip.mismatch() {
		return fmt.Errorf("zip file mismatch")
	}
	return nil
}

type getPathFunc func(BuildConfig) string

func getHashes(primary, verification BuildConfig, getPath func(BuildConfig) string) (FileHashes, error) {
	var fh FileHashes
	var err error
	if fh.Primary, err = digest.FileSHA256Hex(getPath(primary)); err != nil {
		return fh, err
	}
	if fh.Verification, err = digest.FileSHA256Hex(getPath(verification)); err != nil {
		return fh, err
	}
	return fh, nil
}
