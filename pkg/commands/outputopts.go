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

type output struct {
	logOpts
	jsonStdErr bool
	json       bool
}

func (p *output) ReadEnv() error {
	// Write result to stderr by default if not quiet and either verbose or term.
	p.jsonStdErr = !p.logOpts.quietFlag && (p.logOpts.verboseFlag || log.IsVerbose())
	return nil
}

func (p *output) Flags(fs *flag.FlagSet) {
	p.logOpts.Flags(fs)
	p.ownFlags(fs)
}

func (p *output) ownFlags(fs *flag.FlagSet) {
	fs.BoolVar(&p.json, "json", p.json, "print the result json to stdout")
}

type Result interface {
	Error() error
	IsFromCache() bool
}

func (p *output) result(what string, r Result) error {
	// For all failure cases just return an error, which
	// will always be shown to the user.
	if dumped, err := p.maybeDumpJSON(r); err != nil {
		return err
	} else if dumped {
		return r.Error()
	}

	var cached string
	if r.IsFromCache() {
		cached = fmt.Sprintf(" (cached)")
	}
	if err := r.Error(); err != nil {
		return fmt.Errorf("%s failed%s: %w", what, cached, err)
	}
	// For the success case, log immediately.
	p.loud("%s succeeded%s; use the -json flag to see the full result.", what, cached)
	return nil
}

func (p *output) maybeDumpJSON(v any) (bool, error) {
	if p.json {
		return true, dumpJSON(os.Stdout, v)
	}
	if p.jsonStdErr {
		return true, dumpJSON(os.Stderr, v)
	}
	return false, nil
}

func (p *output) productInfo(product crt.Product) error {
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
