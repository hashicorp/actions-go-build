package crt

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
	tmp "github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/tmptest"
	"github.com/hashicorp/go-version"
)

type versionFile struct {
	path, version string
}

func TestGetCoreVersionFromVersionFile_ok(t *testing.T) {

	cases := []struct {
		desc  string
		want  string
		files []versionFile
	}{
		{
			desc:  "no version files",
			want:  "0.0.0-version-file-missing",
			files: nil,
		},
		{
			desc: "no version files in correct paths",
			want: "0.0.0-version-file-missing",
			files: []versionFile{
				{"random/path/VERSION", "1.2.3"},
			},
		},
		{
			desc: "standard",
			want: "1.2.3",
			files: []versionFile{
				{".release/VERSION", "1.2.3"},
			},
		},
		{
			desc: "root",
			want: "1.2.3",
			files: []versionFile{
				{"VERSION", "1.2.3"},
			},
		},
		{
			desc: "dev",
			want: "1.2.3",
			files: []versionFile{
				{"dev/VERSION", "1.2.3"},
			},
		},
		{
			desc: "root beats .release",
			want: "1.2.3-root",
			files: []versionFile{
				{".release/VERSION", "1.2.3-release"},
				{"VERSION", "1.2.3-root"},
			},
		},
		{
			desc: ".release beats dev",
			want: "1.2.3-release",
			files: []versionFile{
				{".release/VERSION", "1.2.3-release"},
				{"dev/VERSION", "1.2.3-dev"},
			},
		},
	}

	for _, c := range cases {
		desc, files, want := c.desc, c.files, version.Must(version.NewVersion(c.want))
		t.Run(desc, func(t *testing.T) {
			// Setup.
			dir := writeTmpFileTree(t, files)
			// Run.
			got, err := getCoreVersionFromVersionFile(dir)
			if err != nil {
				t.Fatal(err)
			}
			// Assert.
			assert.Equal(t, got, want)
		})
	}
}

func TestGetCoreVersionFromVersionFile_err(t *testing.T) {

	cases := []struct {
		desc  string
		want  string
		files []versionFile
	}{
		{
			desc: "invalid version",
			want: `parsing version file ".release/VERSION": invalid version "this isn't a version"`,
			files: []versionFile{
				{".release/VERSION", "this isn't a version"},
			},
		},
		{
			desc: "version contains metadata",
			want: `parsing version file "VERSION": version "1.2.3-release+ent" contains metadata`,
			files: []versionFile{
				{"VERSION", "1.2.3-release+ent"},
			},
		},
	}

	for _, c := range cases {
		desc, files, want := c.desc, c.files, c.want
		t.Run(desc, func(t *testing.T) {
			// Setup.
			dir := writeTmpFileTree(t, files)
			// Run.
			_, gotErr := getCoreVersionFromVersionFile(dir)
			// Assert.
			if gotErr == nil {
				t.Fatalf("got nil error; want error containing %q", want)
			}
			got := gotErr.Error()
			if !strings.Contains(got, want) {
				t.Errorf("got %q; want it to contain %q", got, want)
			}
		})
	}
}

func writeTmpFileTree(t *testing.T, files []versionFile) string {
	t.Helper()
	dir := tmp.Dir(t)
	for _, f := range files {
		p := filepath.Join(dir, f.path)
		if err := fs.WriteFile(p, f.version); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
