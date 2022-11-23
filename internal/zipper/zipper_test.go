package zipper

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
)

type files map[string]string

func TestZipper_ZipDir_ok(t *testing.T) {

	cases := []files{
		{
			"text.txt": "hello!",
		},
	}

	for _, c := range cases {
		inputFiles := c
		t.Run("", func(t *testing.T) {
			buf := &bytes.Buffer{}
			z := New(buf, t.Logf)

			dir := createTestDir(t, inputFiles)
			if err := z.ZipDir(dir); err != nil {
				t.Fatal(err)
			}

			byteReader := bytes.NewReader(buf.Bytes())
			reader, err := zip.NewReader(byteReader, int64(buf.Len()))
			if err != nil {
				t.Fatal(err)
			}

			wantFiles := make(files, len(inputFiles))
			for name := range inputFiles {
				// We expect to see a flattened hierarchy here,
				// so throw away the dir component.
				wantFiles[filepath.Base(name)] = ""
			}

			for _, f := range reader.File {
				_, ok := wantFiles[f.Name]
				if !ok {
					t.Errorf("unexpected file %q", f.Name)
					continue
				}
				delete(wantFiles, f.Name)
				t.Logf("Zip contains %q", f.Name)
			}

			for name := range wantFiles {
				t.Errorf("file %q missing from zip", name)
			}

		})
	}
}

func createTestDir(t *testing.T, f files) string {
	t.Helper()
	pathSegments := strings.Split(t.Name(), "/")
	dir, err := os.MkdirTemp("", fmt.Sprintf("%s.*", pathSegments[0]))
	if len(pathSegments) > 1 {
		dir = filepath.Join(append([]string{dir}, pathSegments...)...)
		if err := fs.Mkdir(dir); err != nil {
			t.Fatal(err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	for name, contents := range f {
		filePathSegments := append([]string{dir}, strings.Split(name, "/")...)
		path := filepath.Join(filePathSegments...)
		if err := fs.Mkdir(filepath.Dir(path)); err != nil {
			t.Fatal(err)
		}
		if err := fs.WriteFile(path, contents); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
