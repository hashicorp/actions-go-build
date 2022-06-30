package fs

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Bytes interface {
	[]byte | string
}

// WriteFile writes a file to the specified path, with default permissions, and
// creates any needed directories.
func WriteFile[T Bytes](path string, contents T) error {
	if err := Mkdir(filepath.Dir(path)); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(contents), os.ModePerm)
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
