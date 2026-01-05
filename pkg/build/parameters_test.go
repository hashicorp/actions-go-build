// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"testing"
)

func TestParseGoVersion(t *testing.T) {

	cases := []struct {
		in   string
		want string
	}{
		{"1.18", "1.18"},
		{"1.18\n", "1.18"},
		{"v1.18", "1.18"},
		{"v1.18\n", "1.18"},
		{"go1.18", "1.18"},
		{"go1.18\n", "1.18"},
		{"1.24", "1.24"},
		{"1.24\n", "1.24"},
		{"v1.24", "1.24"},
		{"v1.24\n", "1.24"},
		{"go1.24", "1.24"},
		{"go1.24\n", "1.24"},
	}

	for _, c := range cases {
		in, want := c.in, c.want
		t.Run(in, func(t *testing.T) {
			got := parseGoVersion(in)
			if got != want {
				t.Errorf("got %q=>%q; want %q", in, got, want)
			}
		})
	}

}
