package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileExists returns a boolean indicating that name is a real path
// and is not a directory.
func FileExists(name string) (bool, error) {
	return existsAndPassesTest(name, func(info os.FileInfo) bool {
		return !info.IsDir()
	})
}

// WriteFile writes a file to the specified path, with default permissions, and
// creates any needed directories.
func WriteFile[T Bytes](path string, contents T) error {
	if err := Mkdir(filepath.Dir(path)); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(contents), os.ModePerm)
}
