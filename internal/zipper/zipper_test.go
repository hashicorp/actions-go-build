package zipper

import (
	"bytes"
	"testing"
)

func TestZipper_writeEntry(t *testing.T) {

	buf := &bytes.Buffer{}

	z := New(buf)

	source := bytes.NewBufferString("blah")

	if err := z.writeEntry("test", source); err != nil {
		t.Fatal(err)
	}

	// Second write of same entry should fail.
	want := `duplicate entry "test"`
	gotErr := z.writeEntry("test", source)
	if gotErr == nil {
		t.Fatalf("got nil error; want %q", want)
	}
	got := gotErr.Error()
	if got != want {
		t.Errorf("got error %q; want %q", got, want)
	}
}
