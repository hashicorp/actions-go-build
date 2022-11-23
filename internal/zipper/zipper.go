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
	log     func(string, ...any)
}

// New returns a new zipper configured to zip the contents of dir.
func New(w io.Writer, logFunc func(string, ...any)) *Zipper {
	return &Zipper{
		written: map[string]struct{}{},
		zw:      zip.NewWriter(w),
		log:     logFunc,
	}
}

// ZipDir zips the file contents of dir to the provided writer, and ignores files
// contained in subdirectories.
// It aims to produce reproducible zips by writing entries in a predictable
// order.
func (z *Zipper) ZipDir(dir string) error {
	z.log("Zipping %q", dir)
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			// We want to scan inside the root dir only, so
			// return null to allow us to keep walking.
			return nil
		}
		if d.IsDir() {
			// Any other dirs are skipped entirely.
			return fs.SkipDir
		}
		name := filepath.Base(path)

		z.log("Adding %q to zip file, from %q", name, path)

		source, err := os.Open(path)
		if err != nil {
			return err
		}
		var closeErr error
		defer func() {
			if err := source.Close(); err != nil {
				closeErr = err
			}
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

	z.log("Finished zipping %q", dir)

	return nil
}

func (z *Zipper) writeEntry(name string, source *os.File) error {
	info, err := source.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	if _, ok := z.written[header.Name]; ok {
		return fmt.Errorf("duplicate entry %q", name)
	}

	header.Method = zip.Deflate

	z.written[name] = struct{}{}
	entry, err := z.zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, source)
	return err
}

// ZipToFile is a convenience function that zips straight to a file.
func ZipToFile(dir, zipFile string, logFunc func(string, ...any)) error {
	f, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = f.Close() }()

	z := New(f, logFunc)

	if err := z.ZipDir(dir); err != nil {
		return err
	}

	return closeErr
}
