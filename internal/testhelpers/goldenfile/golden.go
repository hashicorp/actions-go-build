package goldenfile

import (
	"flag"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update golden files")

func new(t *testing.T) *GF {
	flag.Parse()
	return &GF{
		t:              t,
		actualFile:     nil,
		goldenFileName: filepath.Join("testdata", t.Name()+".golden"),
		actualFileName: t.Name() + ".actual",
	}
}

type GF struct {
	t              *testing.T
	actualFile     *os.File
	goldenFileName string
	actualFileName string
}

// FileAction is a function that writes to the 'got' file.
type FileAction func(got *os.File)

// Do allows you to run a function to write to the actual file.
// It asserts that the actual file written matches the golden file.
func Do(t *testing.T, fn FileAction) {
	t.Helper()
	gf := new(t)
	gf.createActual()
	defer gf.clean()
	fn(gf.actualFile)
	got := gf.readActual()
	if *update {
		gf.update()
	}
	want := gf.read()
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func (gf *GF) read() string {
	gf.t.Helper()
	readBytes, err := ioutil.ReadFile(gf.goldenFileName)
	if err != nil {
		gf.t.Fatal(err)
	}
	return string(readBytes)
}

func (gf *GF) update() {
	gf.t.Helper()
	file := gf.goldenFileName
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
		gf.t.Fatal(err)
	}
	contents, err := ioutil.ReadFile(gf.actualFile.Name())
	if err != nil {
		gf.t.Fatal(err)
	}
	if err := ioutil.WriteFile(file, contents, fs.ModePerm); err != nil {
		gf.t.Fatal(err)
	}
}

func (gf *GF) createActual() *os.File {
	gf.t.Helper()
	name := gf.actualFileName
	f, err := os.CreateTemp("", name+".*")
	if err != nil {
		gf.t.Fatal(err)
	}
	gf.actualFile = f
	return f
}

func (gf *GF) readActual() string {
	gf.t.Helper()
	if gf.actualFile == nil {
		gf.t.Fatal("CreateActual() must be called before ReadActual()")
	}
	readBytes, err := ioutil.ReadAll(gf.actualFile)
	if err != nil {
		gf.t.Fatal(err)
	}
	return string(readBytes)
}

func (gf *GF) clean() {
	gf.t.Helper()
	if gf.actualFile == nil {
		return
	}
	if err := gf.actualFile.Close(); err != nil {
		gf.t.Error(err)
	}
	if err := os.Remove(gf.actualFile.Name()); err != nil {
		gf.t.Error(err)
	}
}
