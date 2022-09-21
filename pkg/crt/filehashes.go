package crt

import (
	"fmt"
	"path/filepath"
)

type FileHashes struct {
	Name, Description string
	SHA256            HashPair
}

func NewFileHashes(desc, primaryPath, verificationPath string) (FileHashes, error) {
	fh := FileHashes{Description: desc}
	pName, vName := filepath.Base(primaryPath), filepath.Base(verificationPath)
	if pName != vName {
		return fh, fmt.Errorf("primary and verification filenames do not match: %q and %q respectively", pName, vName)
	}
	fh.Name = pName
	var err error
	fh.SHA256, err = NewHashPair(primaryPath, verificationPath)
	return fh, err
}

func (fh FileHashes) mismatch() bool {
	return fh.SHA256.mismatch()
}
