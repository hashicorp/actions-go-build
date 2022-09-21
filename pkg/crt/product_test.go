package crt

import (
	"testing"
	"time"

	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
	"github.com/hashicorp/go-version"
)

func TestProduct_Init(t *testing.T) {

	cases := []struct {
		desc    string
		initial Product
		rc      RepoContext
		want    Product
	}{
		{
			"blank product",
			Product{},
			testRepoContext(),
			testProduct(),
		},
		{
			"full version set",
			Product{
				Version: ProductVersion{
					Full: "2.0.0",
				},
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Version = ProductVersion{
					Full: "2.0.0",
					Core: "2.0.0",
					Meta: "",
				}
			}),
		},
		{
			"full version with meta set",
			Product{
				Version: ProductVersion{
					Full: "2.0.0+meta.1",
				},
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Version = ProductVersion{
					Full: "2.0.0+meta.1",
					Core: "2.0.0",
					Meta: "meta.1",
				}
			}),
		},
		{
			"core version found - version not set",
			Product{},
			testRepoContext(func(rc *RepoContext) {
				rc.CoreVersion = *version.Must(version.NewVersion("2.0.0"))
			}),
			testProduct(func(p *Product) {
				p.Version = ProductVersion{
					Full: "2.0.0",
					Core: "2.0.0",
					Meta: "",
				}
			}),
		},
		{
			"core version found - version meta set - version not set",
			Product{
				Version: ProductVersion{
					Meta: "ent",
				},
			},
			testRepoContext(func(rc *RepoContext) {
				rc.CoreVersion = *version.Must(version.NewVersion("2.0.0"))
			}),
			testProduct(func(p *Product) {
				p.Version.Full = "2.0.0+ent"
				p.Version.Core = "2.0.0"
				p.Version.Meta = "ent"
			}),
		},
		{
			"name set",
			Product{
				Name: "blargle",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Name = "blargle"
				p.CoreName = "blargle"
				p.ExecutableName = "blargle"
			}),
		},
		{
			"repository set",
			Product{
				Repository: "othercorp/blargle",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Repository = "othercorp/blargle"
				p.Name = "blargle"
				p.CoreName = "blargle"
				p.ExecutableName = "blargle"
			}),
		},
		{
			"repository and name set",
			Product{
				Repository: "othercorp/blargle",
				Name:       "fish",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Repository = "othercorp/blargle"
				p.Name = "fish"
				p.CoreName = "fish"
				p.ExecutableName = "fish"
			}),
		},
		{
			"repository set enterprise",
			Product{
				Repository: "othercorp/blargle-enterprise",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Repository = "othercorp/blargle-enterprise"
				p.Name = "blargle-enterprise"
				p.CoreName = "blargle"
				p.ExecutableName = "blargle"
			}),
		},
		{
			"name set enterprise",
			Product{
				Name: "shipit-enterprise",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Name = "shipit-enterprise"
				p.CoreName = "shipit"
				p.ExecutableName = "shipit"
			}),
		},
		{
			"repository and name set enterprise",
			Product{
				Repository: "othercorp/blargle-enterprise",
				Name:       "fish-enterprise",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Repository = "othercorp/blargle-enterprise"
				p.Name = "fish-enterprise"
				p.CoreName = "fish"
				p.ExecutableName = "fish"
			}),
		},
		{
			"product in subdirectory",
			Product{},
			testRepoContext(func(rc *RepoContext) {
				rc.RootDir = "/test"
				rc.Dir = "/test/subdir"
			}),
			testProduct(func(p *Product) {
				p.Name = "subdir"
				p.CoreName = "subdir"
				p.ExecutableName = "subdir"
			}),
		},
		{
			"product in subdirectory enterprise",
			Product{},
			testRepoContext(func(rc *RepoContext) {
				rc.RootDir = "/test"
				rc.Dir = "/test/subdir-enterprise"
			}),
			testProduct(func(p *Product) {
				p.Name = "subdir-enterprise"
				p.CoreName = "subdir"
				p.ExecutableName = "subdir"
			}),
		},
		{
			"executable name overridden",
			Product{
				ExecutableName: "overridden",
			},
			testRepoContext(func(rc *RepoContext) {
				rc.RootDir = "/test"
				rc.Dir = "/test/subdir-enterprise"
			}),
			testProduct(func(p *Product) {
				p.Name = "subdir-enterprise"
				p.CoreName = "subdir"
				p.ExecutableName = "overridden"
			}),
		},
	}

	for _, c := range cases {
		desc, initial, rc, want := c.desc, c.initial, c.rc, c.want
		t.Run(desc, func(t *testing.T) {
			got, err := initial.Init(rc)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, got, want)
		})
	}

}

func testRepoContext(modifiers ...func(*RepoContext)) RepoContext {
	rc := standardRepoContext()
	for _, m := range modifiers {
		m(&rc)
	}
	return rc
}

func standardRepoContext() RepoContext {
	v := version.Must(version.NewVersion("1.2.3"))
	return RepoContext{
		RepoName:    "dadgarcorp/lockbox",
		Dir:         "/root-dir",
		RootDir:     "/root-dir",
		CommitSHA:   "cabba9e",
		CommitTime:  time.Date(2022, 7, 13, 12, 50, 1, 1, time.UTC),
		CoreVersion: *v,
	}
}

func testProduct(modifiers ...func(*Product)) Product {
	p := standardProduct()
	for _, m := range modifiers {
		m(&p)
	}
	return p
}

func standardProduct() Product {
	return Product{
		Repository:     "dadgarcorp/lockbox",
		Name:           "lockbox",
		CoreName:       "lockbox",
		ExecutableName: "lockbox",
		Version: ProductVersion{
			Full: "1.2.3",
			Core: "1.2.3",
			Meta: "",
		},
		Revision:     "cabba9e",
		RevisionTime: "2022-07-13T12:50:01Z",
	}
}

func TestProduct_VersionCommandOutput(t *testing.T) {
	cases := []struct {
		desc      string
		in        Product
		wantLong  string
		wantShort string
	}{
		{
			"clean",
			Product{
				Name: "lockbox",
				Version: ProductVersion{
					Full: "1.2.3-beta+meta",
				},
				SourceHash:   "18f94bdcebddbf044de219f8586b054aa3ef0ed3",
				Revision:     "18f94bdcebddbf044de219f8586b054aa3ef0ed3",
				RevisionTime: "2022-08-05T12:39:04Z",
			},
			"lockbox v1.2.3-beta+meta (18f94bdcebddbf044de219f8586b054aa3ef0ed3) 2022-08-05T12:39:04Z",
			"lockbox v1.2.3-beta+meta (18f94bdc) 2022-08-05",
		},
		{
			"dirty",
			Product{
				Name: "lockbox",
				Version: ProductVersion{
					Full: "1.2.3-beta+meta",
				},
				SourceHash:   "blah",
				Revision:     "18f94bdcebddbf044de219f8586b054aa3ef0ed3",
				RevisionTime: "2022-08-05T12:39:04Z",
			},
			"Dirty build: source hash: blah\n" +
				"lockbox v1.2.3-beta+meta (18f94bdcebddbf044de219f8586b054aa3ef0ed3) 2022-08-05T12:39:04Z",
			"Dirty build: source hash: blah\n" +
				"lockbox v1.2.3-beta+meta (18f94bdc) 2022-08-05",
		},
	}

	for _, c := range cases {
		in, wantLong, wantShort := c.in, c.wantLong, c.wantShort
		t.Run(wantShort, func(t *testing.T) {
			gotLong := in.VersionCommandOutput()
			if gotLong != wantLong {
				t.Errorf("got long:\n\t%q\nwant:\n\t%q", gotLong, wantLong)
			}
			gotShort := in.VersionCommandOutputShort()
			if gotShort != wantShort {
				t.Errorf("got short:\n\t%q\nwant:\n\t%q", gotShort, wantShort)
			}
		})
	}
}
