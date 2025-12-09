// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

func GetAllHashes(primary, verification Config) (crt.FileSetHashes, error) {
	getBinPath := func(bc Config) string { return bc.Paths.BinPath }
	getZipPath := func(bc Config) string { return bc.Paths.ZipPath }

	var err error
	var bin, zip crt.FileHashes

	if bin, err = getHashes("executable", primary, verification, getBinPath); err != nil {
		return crt.FileSetHashes{}, err
	}

	if zip, err = getHashes("zip", primary, verification, getZipPath); err != nil {
		return crt.FileSetHashes{}, err
	}

	return crt.NewFileSetHashes(bin, zip), nil

}

type getPathFunc func(Config) string

func getHashes(desc string, primary, verification Config, getPath func(Config) string) (crt.FileHashes, error) {
	pPath, vPath := getPath(primary), getPath(verification)
	return crt.NewFileHashes(desc, pPath, vPath)
}
