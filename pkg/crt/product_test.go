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
			"version set",
			Product{
				Version: "2.0.0",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Version = "2.0.0"
				p.CoreVersion = "2.0.0"
			}),
		},
		{
			"version with meta set",
			Product{
				Version: "2.0.0+meta.1",
			},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Version = "2.0.0+meta.1"
				p.CoreVersion = "2.0.0"
				p.VersionMeta = "meta.1"
			}),
		},
		{
			"core version found - version not set",
			Product{},
			testRepoContext(func(rc *RepoContext) {
				rc.CoreVersion = *version.Must(version.NewVersion("2.0.0"))
			}),
			testProduct(func(p *Product) {
				p.Version = "2.0.0"
				p.CoreVersion = "2.0.0"
			}),
		},
		{
			"core version found - version meta set - version not set",
			Product{
				VersionMeta: "ent",
			},
			testRepoContext(func(rc *RepoContext) {
				rc.CoreVersion = *version.Must(version.NewVersion("2.0.0"))
			}),
			testProduct(func(p *Product) {
				p.Version = "2.0.0+ent"
				p.CoreVersion = "2.0.0"
				p.VersionMeta = "ent"
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
			}),
		},
	}

	for _, c := range cases {
		desc, initial, rc, want := c.desc, c.initial, c.rc, c.want
		t.Run(desc, func(t *testing.T) {
			got := initial.Init(rc)
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
		Dir:         "not relevant to this test",
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
		Repository:   "dadgarcorp/lockbox",
		Name:         "lockbox",
		CoreName:     "lockbox",
		Version:      "1.2.3",
		CoreVersion:  "1.2.3",
		VersionMeta:  "",
		Revision:     "cabba9e",
		RevisionTime: "2022-07-13T12:50:01Z",
	}
}
