package build

import (
	"testing"
)

func TestParseGoVersion(t *testing.T) {

	cases := []struct {
		in   string
		want string
	}{
		{"1.14", "1.14"},
		{"1.14\n", "1.14"},
		{"v1.14", "1.14"},
		{"v1.14\n", "1.14"},
		{"go1.14", "1.14"},
		{"go1.14\n", "1.14"},
		{"1.18", "1.18"},
		{"1.18\n", "1.18"},
		{"v1.18", "1.18"},
		{"v1.18\n", "1.18"},
		{"go1.18", "1.18"},
		{"go1.18\n", "1.18"},
	}

	for _, c := range cases {
		in, want := c.in, c.want
		t.Run(in, func(t *testing.T) {
			got := parseGoVersion(in)
			if got != want {
				t.Errorf("got %q=>%q; want %q", in, got, want)
			}
		})
	}

}
