package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type outputOpts struct {
	logOpts
	jsonStdErr bool
	json       bool
}

func (p *outputOpts) ReadEnv() error {
	// Write result to stderr by default if not quiet and either verbose or term.
	p.jsonStdErr = !p.logOpts.quietFlag && (p.logOpts.verboseFlag || log.IsVerbose())
	return nil
}

func (p *outputOpts) Flags(fs *flag.FlagSet) {
	p.logOpts.Flags(fs)
	p.ownFlags(fs)
}

func (p *outputOpts) ownFlags(fs *flag.FlagSet) {
	fs.BoolVar(&p.json, "json", p.json, "print the result json to stdout")
}

type Result interface {
	Error() error
}

func (p *outputOpts) result(what string, r Result) error {
	resultErr := r.Error()
	dumped, err := p.maybeDumpJSON(r)
	if err != nil {
		return err
	}
	if dumped {
		return resultErr
	}

	resultStatus := "succeeded"
	if resultErr != nil {
		resultStatus = "failed"
	}
	p.loud("%s %s; use the -json flag to see the full result.", what, resultStatus)
	return resultErr
}

func (p *outputOpts) maybeDumpJSON(v any) (bool, error) {
	if p.json {
		return true, dumpJSON(os.Stdout, v)
	}
	if p.jsonStdErr {
		return true, dumpJSON(os.Stderr, v)
	}
	return false, nil
}

func (p *outputOpts) productInfo(product crt.Product) error {
	if dumped, err := p.maybeDumpJSON(product); dumped || err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := dumpJSON(buf, product); err != nil {
		return err
	}
	s := buf.String()
	s = strings.ReplaceAll(s, `",`, "")
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `},`, "")
	s = strings.ReplaceAll(s, `}`, "")
	s = strings.ReplaceAll(s, `{`, "")
	_, err := fmt.Fprint(os.Stdout, s)
	return err
}

func dumpJSON(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}
