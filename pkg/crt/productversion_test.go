// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package crt

import (
	"testing"
)

func TestProductVersion_Init(t *testing.T) {
	pvs := func(pvs ...ProductVersion) []ProductVersion { return pvs }

	// For each case, all the values in "in" should resolve to the same "want".
	cases := []struct {
		desc string
		in   []ProductVersion
		want ProductVersion
	}{
		{
			"empty or default->default",
			pvs(
				ProductVersion{},
				ProductVersion{
					Full: "0.0.0-unversioned+local",
				},
				ProductVersion{
					Full: "0.0.0-unversioned+local",
					Core: "0.0.0-unversioned",
				},
				ProductVersion{
					Full: "0.0.0-unversioned+local",
					Core: "0.0.0-unversioned",
					Meta: "local",
				},
				ProductVersion{
					Core: "0.0.0-unversioned",
					Meta: "local",
				},
				ProductVersion{
					Meta: "local",
				},
			),
			ProductVersion{
				Full: "0.0.0-unversioned+local",
				Core: "0.0.0-unversioned",
				Meta: "local",
			},
		},
		{
			"full->equal",
			pvs(
				ProductVersion{
					Full: "1.2.3-pre+meta",
					Core: "1.2.3-pre",
					Meta: "meta",
				},
				ProductVersion{
					Full: "1.2.3-pre+meta",
					Meta: "meta",
				},
				ProductVersion{
					Full: "1.2.3-pre+meta",
					Core: "1.2.3-pre",
				},
				ProductVersion{
					Full: "1.2.3-pre+meta",
				},
				ProductVersion{
					Core: "1.2.3-pre",
					Meta: "meta",
				},
			),
			ProductVersion{
				Full: "1.2.3-pre+meta",
				Core: "1.2.3-pre",
				Meta: "meta",
			},
		},
		{
			"core->full",
			pvs(
				ProductVersion{
					Core: "4.5.6-pre",
				},
				ProductVersion{
					Full: "4.5.6-pre",
					Core: "4.5.6-pre",
				},
				ProductVersion{
					Full: "4.5.6-pre",
				},
			),
			ProductVersion{
				Full: "4.5.6-pre",
				Core: "4.5.6-pre",
				Meta: "",
			},
		},
		{
			"meta->defaultwithmeta",
			pvs(
				ProductVersion{
					Meta: "hello",
				},
			),
			ProductVersion{
				Full: "0.0.0-unversioned+hello",
				Core: "0.0.0-unversioned",
				Meta: "hello",
			},
		},
	}

	for _, c := range cases {
		desc, in, want := c.desc, c.in, c.want
		t.Run(desc, func(t *testing.T) {
			for _, in := range in {
				t.Run("", func(t *testing.T) {
					got, err := in.Init()
					if err != nil {
						t.Fatal(err)
					}
					if got != want {
						t.Errorf("got:\n% #v;\nwant:\n% #v", got, want)
					}
				})
			}
		})
	}
}
