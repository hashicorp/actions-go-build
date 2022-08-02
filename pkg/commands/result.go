package commands

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/hashicorp/actions-go-build/internal/log"
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

func dumpJSON(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}
