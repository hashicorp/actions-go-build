package fs

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

func DirExists(name string) (bool, error) {
	return existsAndPassesTest(name, func(info os.FileInfo) bool {
		return info.IsDir()
	})
}

// Mkdir makes the directory at path, using default permissions, and logs its activity.
func Mkdir(path string) error {
	log.Printf("Creating directory %q", path)
	return os.MkdirAll(path, fs.ModePerm)
}

// Mkdirs calls Mkdir sequentially on paths and returns an error after the first failure.
func Mkdirs(paths ...string) error {
	for _, p := range paths {
		if err := Mkdir(p); err != nil {
			return err
		}
	}
	return nil
}

// SetMtime sets the mtime of all files inside dir to the provided time,
// and logs its activity.
func SetMtimes(dir string, to time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		filePath := filepath.Join(dir, e.Name())
		log.Printf("Updating mtime for %q to %s", filePath, to)
		if err := os.Chtimes(filePath, to, to); err != nil {
			return err
		}
	}
	return nil
}
