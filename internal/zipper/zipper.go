package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Zipper struct {
	dir     string
	written map[string]struct{}
	zw      *zip.Writer
}

// NewZipper returns a new zipper configured to zip the contents of dir.
func New(w io.Writer) *Zipper {
	return &Zipper{
		written: map[string]struct{}{},
		zw:      zip.NewWriter(w),
	}
}

// ZipDir zips the contents of dir to the provided writer, and flattens any
// directory hierarchy, so the resultant zip has just a flat list of
// files. Filename conflicts result in error.
// It aims to produce reproducible zips by writing entries in a predictable
// order.
//
// It is intended to perform the same function as calling 'zip -Xrj $zipFile $dir'
func (z *Zipper) ZipDir(dir string) error {
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		name := filepath.Base(path)
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		var closeErr error
		defer func() {
			closeErr = source.Close()
		}()
		if err := z.writeEntry(name, source); err != nil {
			return err
		}
		return closeErr
	}); err != nil {
		return err
	}
	if err := z.zw.Close(); err != nil {
		return err
	}

	return nil
}

func (z *Zipper) writeEntry(name string, source io.Reader) error {
	if _, ok := z.written[name]; ok {
		return fmt.Errorf("duplicate entry %q", name)
	}
	z.written[name] = struct{}{}
	entry, err := z.zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, source)
	return err
}

// ZipToFile is a convenience function meant to be equivalent to using the command line:
// 'zip -Xrj $zipFile $dir`
func ZipToFile(dir, zipFile string) error {
	f, err := os.OpenFile(zipFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = f.Close() }()

	z := New(f)

	if err := z.ZipDir(dir); err != nil {
		return err
	}

	return closeErr
}
