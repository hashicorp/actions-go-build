// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package crt

import "github.com/hashicorp/actions-go-build/pkg/digest"

type HashPair struct {
	Primary, Verification string
	Match                 bool
}

func NewHashPair(primaryPath, verificationPath string) (HashPair, error) {
	var hp HashPair
	var err error
	if hp.Primary, err = digest.FileSHA256Hex(primaryPath); err != nil {
		return hp, err
	}
	if hp.Verification, err = digest.FileSHA256Hex(verificationPath); err != nil {
		return hp, err
	}
	hp.Match = !hp.mismatch()
	return hp, nil
}

// mismatch returns true if the hashes are different, or if they are both empty.
func (hp HashPair) mismatch() bool {
	return hp.Primary != hp.Verification && hp.Primary != ""
}
