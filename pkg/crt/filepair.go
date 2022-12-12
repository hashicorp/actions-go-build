// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package crt

// FilePair represents a pair of files that should be identical.
// (The primary and verirication build's version of each file.)
type FilePair struct {
	Primary, Verification File
}
