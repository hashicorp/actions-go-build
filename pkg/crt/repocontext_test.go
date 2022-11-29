// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package crt

import (
	"testing"
)

func TestGetRepoNameFromRemoteURL(t *testing.T) {
	// Each case should have the same answer: "dadgarcorp/lockbox"
	cases := []string{
		// Local filesystem
		"/Users/dave/src/github.com/dadgarcorp/lockbox/.",
		"/Users/dave/src/github.com/dadgarcorp/lockbox/",
		"/Users/dave/src/github.com/dadgarcorp/lockbox",
		"../dadgarcorp/lockbox/.",
		"../dadgarcorp/lockbox/",
		"../dadgarcorp/lockbox",
		"../../../dadgarcorp/lockbox/.",
		"../../../dadgarcorp/lockbox/",
		"../../../dadgarcorp/lockbox",

		// Special Git SSH URLs (i.e. not real URLs)
		"git@github.com:dadgarcorp/lockbox.git",
		"git@github.com:dadgarcorp/lockbox",
		"git@github.com:dadgarcorp/lockbox.git/",
		"git@github.com:dadgarcorp/lockbox/",
		"blah@blah.com:dadgarcorp/lockbox.git",
		"blah@blah.com:dadgarcorp/lockbox",
		"blah@blah.com:dadgarcorp/lockbox.git/",
		"blah@blah.com:dadgarcorp/lockbox/",

		// Normal URLs
		"https://github.com/dadgarcorp/lockbox.git",
		"https://github.com/dadgarcorp/lockbox.git/",
		"https://github.com/dadgarcorp/lockbox",
		"https://github.com/dadgarcorp/lockbox/",
		"https://blah.com/dadgarcorp/lockbox.git",
		"https://blah.com/dadgarcorp/lockbox.git/",
		"https://blah.com/dadgarcorp/lockbox",
		"https://blah.com/dadgarcorp/lockbox/",
		"http://github.com/dadgarcorp/lockbox.git",
		"http://github.com/dadgarcorp/lockbox.git/",
		"http://github.com/dadgarcorp/lockbox",
		"http://github.com/dadgarcorp/lockbox/",
		"http://blah.com/dadgarcorp/lockbox.git",
		"http://blah.com/dadgarcorp/lockbox.git/",
		"http://blah.com/dadgarcorp/lockbox",
		"http://blah.com/dadgarcorp/lockbox/",
		"git://github.com/dadgarcorp/lockbox.git",
		"git://github.com/dadgarcorp/lockbox.git/",
		"git://github.com/dadgarcorp/lockbox",
		"git://github.com/dadgarcorp/lockbox/",
	}

	const want = "dadgarcorp/lockbox"

	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			got, err := getRepoNameFromRemoteURL(c)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got %q; want %q", got, want)
			}
		})
	}
}
