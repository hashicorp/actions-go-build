package cli

import (
	"io"
	"text/tabwriter"
)

func makeTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
}

func TabWrite[T any](w io.Writer, data []T, printer func(datum T) string) error {
	tw := makeTabWriter(w)
	for _, d := range data {
		s := printer(d) + "\n"
		if _, err := tw.Write([]byte(s)); err != nil {
			return err
		}
	}
	return tw.Flush()
}
