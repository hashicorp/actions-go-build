// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

//go:embed templates/stepsummary.md.tmpl
var stepSummaryTemplate string

type verifyOpts struct {
	verifyish
	outFile     string
	stepSummary string
}

func (opts *verifyOpts) Flags(fs *flag.FlagSet) {
	opts.verifyish.Flags(fs)
	fs.StringVar(&opts.outFile, "o", "", "write the result json to this file")
	fs.StringVar(&opts.stepSummary, "github-step-summary", os.Getenv("GITHUB_STEP_SUMMARY"), "write a github step summary to this file")
}

var Verify = cli.LeafCommand("verify", "verify a build's reproducibility", func(opts *verifyOpts) error {
	result, err := opts.runVerification()
	if err != nil {
		return err
	}
	if opts.outFile != "" {
		if err := json.WriteFile(opts.outFile, result); err != nil {
			return err
		}
		opts.log("Result written to %s", opts.outFile)
	}
	if opts.stepSummary != "" {
		opts.log("Writing GitHub Step Summary to %s", opts.stepSummary)
		f, err := fs.Append(opts.stepSummary)
		if err != nil {
			return err
		}
		defer f.Close()
		funcs := template.FuncMap{
			"json": func(a any) string {
				s, err := json.String(a)
				if err != nil {
					return fmt.Sprintf("<error: %v>", err)
				}
				return s
			},
		}
		if err := template.Must(template.New("").Funcs(funcs).Parse(stepSummaryTemplate)).Execute(f, result); err != nil {
			return err
		}
	}
	return opts.output.result("Reproducibility verification", result)
})
