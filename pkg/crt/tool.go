// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package crt

// Tool represents a build tool involved in building or analysing build results.
type Tool struct {
	Name, Version, Revision, RevisionTime string
}
