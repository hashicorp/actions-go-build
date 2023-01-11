// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package commands

import (
	"os"
	"testing"

	"github.com/hashicorp/actions-go-build/pkg/build"
)

func test[T any](t *testing.T, desc string, setup func(*buildish), assert func(T, build.Config) error) {
	t.Helper()
	t.Run(desc, func(t *testing.T) {
		b := buildish{}
		setup(&b)
		m, err := b.build("testing")
		if err != nil {
			t.Fatal(err)
		}
		if m == nil {
			t.Fatalf("got nil build manager")
		}

		got := m.Build()
		if got == nil {
			t.Fatalf("got nil build")
		}

		gotConfig := got.Config()

		pb, ok := got.(T)
		if !ok {
			t.Fatalf("got a %T; want a %T", got, pb)
		}

		if err := assert(pb, gotConfig); err != nil {
			t.Error(err)
		}
	})
}

func TestBuildish_build(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	assertNonPWDRoot := func(t *testing.T, b build.Build) {
		t.Helper()
		c := b.Config()
		if c.Paths.WorkDir == wd {
			t.Errorf("got workdir %q; want it to be in verification root", c.Paths.WorkDir)
		}
	}

	assertPWDRoot := func(t *testing.T, b build.Build) {
		t.Helper()
		c := b.Config()
		if c.Paths.WorkDir != wd {
			t.Errorf("got workdir %q; want %q", c.Paths.WorkDir, wd)
		}
	}

	assertVerificationBuild := func(t *testing.T, b build.Build) {
		t.Helper()
		if !b.IsVerification() {
			t.Errorf("got a primary build; want a verification build")
		}
	}

	assertPrimaryBuild := func(t *testing.T, b build.Build) {
		t.Helper()
		if b.IsVerification() {
			t.Errorf("got a verification build; want a primary build")
		}
	}

	test(t, "default/local primary",
		func(b *buildish) {
			// blank
		},
		func(pb *build.Primary, c build.Config) error {
			assertPWDRoot(t, pb)
			assertPrimaryBuild(t, pb)
			return nil
		})

	test(t, "force verification/local verification",
		func(b *buildish) {
			b.buildFlags.forceVerification = true
		},
		func(pb *build.LocalVerification, c build.Config) error {
			assertNonPWDRoot(t, pb)
			assertVerificationBuild(t, pb)
			return nil
		})

	test(t, "buildresult target/remote verification",
		func(b *buildish) {
			b.target = "testdata/example.buildresult.json"
		},
		func(pb *build.RemoteBuild, c build.Config) error {
			assertNonPWDRoot(t, pb)
			assertPrimaryBuild(t, pb)
			return nil
		})

	test(t, "buildresult target/clean",
		func(b *buildish) {
			b.target = "testdata/example.buildresult.json"
		},
		func(pb *build.RemoteBuild, c build.Config) error {
			assertNonPWDRoot(t, pb)
			assertPrimaryBuild(t, pb)
			return nil
		})

}
