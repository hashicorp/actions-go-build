package build

import (
	"testing"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

func TestNewConfig(t *testing.T) {

	// inputs just groups together the inputs to the function
	// for testing convenience.
	type inputs struct {
		product      crt.Product
		params       Parameters
		paths        Paths
		tool         crt.Tool
		reproducible bool
	}

	emptyInputs := func() inputs {
		return inputs{
			product:      crt.Product{},
			params:       Parameters{},
			paths:        Paths{},
			tool:         crt.Tool{},
			reproducible: false,
		}
	}

	cases := []struct {
		name   string
		setup  func(*inputs)
		assert func(*testing.T, Config)
	}{
		{
			name:   "empty",
			setup:  func(*inputs) {},
			assert: func(t *testing.T, c Config) {},
		},
		{
			name: "linux",
			setup: func(i *inputs) {
				i.params.OS = "linux"
				i.product.ExecutableName = "blah"
			},
			assert: func(t *testing.T, c Config) {
				got, want := c.Product.ExecutableName, "blah"
				if c.Product.ExecutableName != want {
					t.Errorf("got executable name %q; want %q", got, want)
				}
			},
		},
		{
			name: "windows no exe",
			setup: func(i *inputs) {
				i.params.OS = "windows"
				i.product.ExecutableName = "blah"
			},
			assert: func(t *testing.T, c Config) {
				got, want := c.Product.ExecutableName, "blah.exe"
				if c.Product.ExecutableName != want {
					t.Errorf("got executable name %q; want %q", got, want)
				}
			},
		},
		{
			name: "windows with exe",
			setup: func(i *inputs) {
				i.params.OS = "windows"
				i.product.ExecutableName = "blah.exe"
			},
			assert: func(t *testing.T, c Config) {
				got, want := c.Product.ExecutableName, "blah.exe"
				if c.Product.ExecutableName != want {
					t.Errorf("got executable name %q; want %q", got, want)
				}
			},
		},
	}

	for _, c := range cases {
		name, setup, assert := c.name, c.setup, c.assert
		t.Run(name, func(t *testing.T) {
			in := emptyInputs()
			setup(&in)
			got, err := NewConfig(in.product, in.params, in.paths, in.tool, in.reproducible)
			if err != nil {
				t.Fatal(err)
			}
			assert(t, got)
		})
	}

}
