// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"
)

var (
	Version      string
	Revision     string
	RevisionTime string
)

func main() {
	fmt.Println("Example app.")

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)

	fmt.Fprintf(w, "Version:\t%s\n", Version)
	fmt.Fprintf(w, "Revision:\t%s\n", Revision)
	fmt.Fprintf(w, "RevisionTime:\t%s\n", RevisionTime)
	fmt.Fprintf(w, "GoVersion:\t%s\n", runtime.Version())
	w.Flush()

	fmt.Println("Bye!")
}
