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

type presenter struct {
	json bool
}

func (p *presenter) ReadEnv() error {
	// If we're not a terminal (e.g. in CI) then default json mode to on.
	p.json = !log.IsTerm()
	return nil
}

func (p *presenter) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&p.json, "json", p.json, "show the result as json")
}

type Result interface {
	Error() error
}

func (p *presenter) result(what string, r Result) error {

	// In JSON mode don't write log messages, just stick to pure JSON output.
	if p.json {
		if err := dumpJSON(os.Stdout, r); err != nil {
			return err
		}
		return r.Error()
	}

	// Otherwise tell the user how to see the full result.
	if err := r.Error(); err != nil {
		log.Info("%s failed; use the -json flag to see the full result.", what)
		return err
	}
	log.Info("%s succeeded; use the -json flag to see the full result.", what)
	return nil
}

func (p *presenter) productInfo(product crt.Product) error {
	if p.json {
		return dumpJSON(os.Stdout, product)
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
