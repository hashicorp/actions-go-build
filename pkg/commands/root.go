// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	"os"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// stdout and stderr can be overridden for testing.
var stdout = os.Stdout
var stderr = os.Stderr

// Root is the root command of the whole CLI. It is given the name "go" so that
// when this CLI is incorporated into a parent CLI, the commands within will be
// rooted at "go". E.g. "go-build", "go-build primary", "go-build verification".
var Root = cli.RootCommand("go-build", "go build and related functions", Build, Verify, Config)
