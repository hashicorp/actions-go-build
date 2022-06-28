package config

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

func TestConfig_ExportToGitHubEnv_ok(t *testing.T) {
	flag.Parse()

	c := standardConfig()

	f, err := os.CreateTemp("", "gh_env.*")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	t.Logf("GITHUB_ENV=%q", f.Name())

	os.Setenv("GITHUB_ENV", f.Name())

	c.ExportToGitHubEnv()

	got := readFileToString(t, f.Name())

	if *update {
		writeGoldenFile(t, got)
	}

	want := readGoldenFile(t)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("ExportToGitHubEnv() mismatch (-want +got):\n%s", diff)
	}
}

func readFileToString(t *testing.T, file string) string {
	t.Helper()
	readBytes, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(readBytes)
}

func goldenFileName(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", t.Name()+".golden")
}

func readGoldenFile(t *testing.T) string {
	t.Helper()
	return readFileToString(t, goldenFileName(t))
}

func writeGoldenFile(t *testing.T, contents string) {
	t.Helper()
	file := goldenFileName(t)
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(file, []byte(contents), fs.ModePerm); err != nil {
		t.Fatal(err)
	}
}
