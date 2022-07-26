package crt

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/digest"
)

func GetAllHashes(primary, verification BuildConfig) (FileSetHashes, error) {
	getBinPath := func(bc BuildConfig) string { return bc.Paths.BinPath }
	getZipPath := func(bc BuildConfig) string { return bc.Paths.ZipPath }

	var fsh FileSetHashes
	var err error

	if fsh.Bin, err = getHashes("executable", primary, verification, getBinPath); err != nil {
		return fsh, err
	}

	if fsh.Zip, err = getHashes("zip", primary, verification, getZipPath); err != nil {
		return fsh, err
	}

	return fsh, nil

}

type FileHashes struct {
	Name, Description string
	SHA256            HashPair
}

type HashPair struct {
	Primary, Verification string
}

func (fh FileHashes) mismatch() bool {
	return fh.SHA256.mismatch()
}

// mismatch returns true if the hashes are different, or if they are both empty.
func (hp HashPair) mismatch() bool {
	return hp.Primary != hp.Verification && hp.Primary != ""
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

func getHashes(desc string, primary, verification BuildConfig, getPath func(BuildConfig) string) (FileHashes, error) {
	fh := FileHashes{Description: desc}
	pPath, vPath := getPath(primary), getPath(verification)
	pName, vName := filepath.Base(pPath), filepath.Base(vPath)
	if pName != vName {
		return fh, fmt.Errorf("primary and verification filenames do not match: %q and %q respectively", pName, vName)
	}
	fh.Name = pName
	var err error
	fh.SHA256, err = getHashPair(pPath, vPath)
	return fh, err
}

func getHashPair(primaryPath, verificationPath string) (HashPair, error) {
	var hp HashPair
	var err error
	if hp.Primary, err = digest.FileSHA256Hex(primaryPath); err != nil {
		return hp, err
	}
	if hp.Verification, err = digest.FileSHA256Hex(verificationPath); err != nil {
		return hp, err
	}
	return hp, nil
}
