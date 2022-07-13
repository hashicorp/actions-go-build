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
			}),
		},
		{
			"repository set",
			Product{Repository: "othercorp/blargle"},
			testRepoContext(),
			testProduct(func(p *Product) {
				p.Repository = "othercorp/blargle"
				p.Name = "blargle"
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
		Version:      "1.2.3",
		Revision:     "cabba9e",
		RevisionTime: "2022-07-13T12:50:01Z",
	}
}
