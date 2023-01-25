// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	"fmt"
	"io"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

type printer struct {
	w           io.Writer
	build       build.Build
	printTitles bool
	prefix      string
}

func firstErr(f ...func() error) error {
	for _, ef := range f {
		if err := ef(); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) title(s string) error {
	if !p.printTitles {
		return nil
	}
	_, err := fmt.Fprintln(p.w, s+":")
	return err
}

func (p *printer) line(s string, a ...any) error {
	_, err := fmt.Fprintf(p.w, p.prefix+s+"\n", a...)
	return err
}

func tabWrite[T any](p *printer, list []T, line func(T) string) error {
	return cli.TabWrite(p.w, list, line)
}
