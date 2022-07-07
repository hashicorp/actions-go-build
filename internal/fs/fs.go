package fs

import (
	"io/fs"
	"log"
	"os"
)

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
