package crt

import (
	"fmt"
)

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
