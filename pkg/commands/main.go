package commands

import "github.com/hashicorp/actions-go-build/pkg/cli"

var Main = cli.RootCommand("build", "build and related functions", Inputs, Run, Env)
