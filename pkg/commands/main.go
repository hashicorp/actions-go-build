package commands

import (
	"os"

	"github.com/hashicorp/actions-go-build/pkg/cli"
)

// stdout and stderr can be overridden for testing.
var stdout = os.Stdout
var stderr = os.Stderr

var Main = cli.RootCommand("build", "build and related functions", Inputs, Run, Env)
