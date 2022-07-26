package build

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

func GetAllHashes(primary, verification Config) (crt.FileSetHashes, error) {
	getBinPath := func(bc Config) string { return bc.Paths.BinPath }
	getZipPath := func(bc Config) string { return bc.Paths.ZipPath }

	var fsh crt.FileSetHashes
	var err error

	if fsh.Bin, err = getHashes("executable", primary, verification, getBinPath); err != nil {
		return fsh, err
	}

	if fsh.Zip, err = getHashes("zip", primary, verification, getZipPath); err != nil {
		return fsh, err
	}

	return fsh, nil

}

type getPathFunc func(Config) string

func getHashes(desc string, primary, verification Config, getPath func(Config) string) (crt.FileHashes, error) {
	fh := crt.FileHashes{Description: desc}
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

func getHashPair(primaryPath, verificationPath string) (crt.HashPair, error) {
	var hp crt.HashPair
	var err error
	if hp.Primary, err = digest.FileSHA256Hex(primaryPath); err != nil {
		return hp, err
	}
	if hp.Verification, err = digest.FileSHA256Hex(verificationPath); err != nil {
		return hp, err
	}
	return hp, nil
}
