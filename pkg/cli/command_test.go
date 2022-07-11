package cli

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/actions-go-build/internal/testhelpers/assert"
)

// args constructs a slice like os.Args, setting the first
// arg to the empty string, which represents the command
// name used to call the CLI.
func args(a ...string) []string {
	return append([]string{""}, a...)
}

type testFlags struct {
	flag1, flag2 bool
}

func (o *testFlags) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&o.flag1, "flag1", false, "flag1 desc")
	fs.BoolVar(&o.flag2, "flag2", false, "flag2 desc")
}

type testArgs struct {
	args []string
}

func (a *testArgs) ParseArgs(args []string) error {
	a.args = args
	return nil
}

type testOpts struct {
	testFlags
	testArgs
}

func testCLI() (Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	write := func(a ...any) error {
		s := make([]string, len(a))
		for i, item := range a {
			s[i] = fmt.Sprint(item)
		}
		_, err := buf.WriteString(strings.Join(s, ", "))
		return err
	}
	root := RootCommand("root", "root command",
		LeafCommand("leaf", "leaf command", func(None) error {
			return write("leaf")
		}),
		LeafCommand("leaf2", "leaf command 2", func(flags *testFlags) error {
			return write("leaf2", flags.flag1, flags.flag2)
		}),
		RootCommand("root2", "root command 2",
			LeafCommand("leaf3", "leaf command 3", func(None) error {
				return write("leaf3")
			}),
		),
		LeafCommand("leaf4", "leaf command 4", func(a *testArgs) error {
			return write("leaf4", strings.Join(a.args, ", "))
		}),
		LeafCommand("leaf5", "leaf command 5", func(o *testOpts) error {
			return write("leaf5", o.flag1, o.flag2, strings.Join(o.args, ", "))
		}),
	)
	return root, buf
}

func TestCommand_ok(t *testing.T) {

	cases := []struct {
		args []string
		want string
	}{
		{
			args(),
			"",
		},
		{
			args("leaf"),
			"leaf",
		},
		{
			args("leaf2"),
			"leaf2, false, false",
		},
		{
			args("leaf2", "-flag1"),
			"leaf2, true, false",
		},
		{
			args("leaf2", "-flag2"),
			"leaf2, false, true",
		},
		{
			args("leaf2", "-flag1", "-flag2"),
			"leaf2, true, true",
		},
		{
			args("root2"),
			"",
		},
		{
			args("root2", "leaf3"),
			"leaf3",
		},
		{
			args("leaf4"),
			"leaf4, ",
		},
		{
			args("leaf4", "hello"),
			"leaf4, hello",
		},
		{
			args("leaf4", "hello", "world"),
			"leaf4, hello, world",
		},
		{
			args("leaf5", "hello", "world"),
			"leaf5, false, false, hello, world",
		},
		{
			args("leaf5", "-flag1", "hello", "world"),
			"leaf5, true, false, hello, world",
		},
		{
			args("leaf5", "-flag2", "hello", "world"),
			"leaf5, false, true, hello, world",
		},
		{
			args("leaf5", "-flag1", "-flag2", "hello", "world"),
			"leaf5, true, true, hello, world",
		},
	}

	for _, c := range cases {
		args, want := c.args, c.want
		t.Run("", func(t *testing.T) {
			c, buf := testCLI()
			if err := c.Execute(args); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, buf.String(), want)
		})
	}
}
