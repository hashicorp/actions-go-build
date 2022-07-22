package build

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/git"
	tmp "github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/tmptest"
)

func TestBuild_Run_ok(t *testing.T) {
	dir := tmp.Dir(t)
	t.Logf("Test dir: %q", dir)

	testBuild, err := New(standardConfig(dir))
	if err != nil {
		t.Fatal(err)
	}

	b := testBuild.(*build)
	b.createTestProductRepo(t)
	result := b.Run()
	if err := result.Error(); err != nil {
		t.Fatal(err)
	}
}

func TestBuild_Run_err(t *testing.T) {
	dir := tmp.Dir(t)
	t.Logf("Test dir: %q", dir)

	c := standardConfig(dir)
	c.Parameters.Instructions = "echo 'oh no!'; exit 1"
	testBuild, err := New(c)
	if err != nil {
		t.Fatal(err)
	}

	b := testBuild.(*build)
	b.createTestProductRepo(t)
	result := b.Run()
	gotErr := result.Error()
	want := "running build instructions failed: exit status 1"
	if gotErr == nil {
		t.Fatalf("got nil error; want %q", want)
	}
	got := gotErr.Error()
	if want != got {
		t.Fatalf("got error %q; want %q", got, want)
	}
}

const mainDotGo = `
	package main

	import "fmt"

	func main() {
		fmt.Println("hello, world")
	}
`

const goDotMod = `module github.com/dadgarcorp/lockbox

go 1.18
`

// createTestProductRepo creates a test repo and returns its path.
func (b *build) createTestProductRepo(t *testing.T) {
	b.writeTestFile(t, "main.go", mainDotGo)
	b.writeTestFile(t, "go.mod", goDotMod)
	repo, err := git.Init(b.config.Paths.WorkDir, git.WithAuthor("test", "test@test.com"))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := repo.Commit("initial commit"); err != nil {
		t.Fatal(err)
	}
}

// must is a quick way to fail a test depending on if an error is nil or not.
func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func (b *build) runTestCommand(t *testing.T, name string, args ...string) {
	must(t, b.runCommand(name, args...))
}

func (b *build) writeTestFile(t *testing.T, name, contents string) {
	name = filepath.Join(b.config.Paths.WorkDir, name)
	must(t, fs.WriteFile(name, contents))
}

func standardCommitTime() (ts time.Time, rfc3339 string) {
	ts = time.Date(2022, 7, 4, 11, 33, 33, 0, time.UTC)
	rfc3339 = ts.Format(time.RFC3339)
	return
}

func standardConfig(workDir string) crt.BuildConfig {
	_, revisionTimestampRFC3339 := standardCommitTime()
	return crt.BuildConfig{
		Product: crt.Product{
			Repository: "dadgarcorp/lockbox",
			Name:       "lockbox",
			Version: crt.ProductVersion{
				Full: "1.2.3",
				Core: "1.2.3",
				Meta: "1.2.3",
			},
			Revision:     "cabba9e",
			RevisionTime: revisionTimestampRFC3339,
		},
		Parameters: crt.BuildParameters{
			Instructions: `echo -n "Building '$BIN_PATH'..." && go build -o $BIN_PATH && echo "Done!"`,
			OS:           "linux",
			Arch:         "amd64",
		},
		Paths: crt.BuildPaths{
			WorkDir:   workDir,
			TargetDir: filepath.Join(workDir, "dist"),
			BinPath:   filepath.Join(workDir, "dist", "lockbox"),
			ZipPath:   filepath.Join(workDir, "out", "lockbox_1.2.3_amd64.zip"),
			MetaDir:   filepath.Join(workDir, "meta"),
		},
	}
}
