package assert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal(t *testing.T, got, want interface{}) {
	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("Mismatch (-got +want):\n%s", diff)
	}
}
