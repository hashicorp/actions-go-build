// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package digest

import "fmt"

func ID[T any](thing T) string {
	zero := *(new(T))
	zeroHash, err := JSONSHA256Hex(zero)
	if err != nil {
		panic(err)
	}
	thingHash, err := JSONSHA256Hex(thing)
	if err != nil {
		panic(err)
	}
	if thingHash == zeroHash {
		panic(fmt.Errorf("Can't take ID of zero %T: % #v", thing, thing))
	}
	return thingHash
}

func CompoundID(things ...any) string {
	ids := make([]string, len(things))
	for i, t := range things {
		ids[i] = ID(t)
	}
	s, err := SHA256HexStrings(ids...)
	if err != nil {
		panic(err)
	}
	return s
}
