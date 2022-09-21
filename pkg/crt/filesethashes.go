package crt

import (
	"fmt"
)

type FileSetHashes struct {
	Bin      FileHashes
	Zip      FileHashes
	AllMatch bool
}

func NewFileSetHashes(bin, zip FileHashes) FileSetHashes {
	return FileSetHashes{
		Bin:      bin,
		Zip:      zip,
		AllMatch: !bin.mismatch() && !zip.mismatch(),
	}
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
