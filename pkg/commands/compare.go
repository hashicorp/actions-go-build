package commands

import (
	"fmt"
	"log"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

var Compare = cli.LeafCommand("compare", "compare digests", func(configs *opts.AllBuildConfigs) error {
	bin, zip := "executable", "zip"
	if err := comp(bin, configs, func(bc crt.BuildConfig) string { return bc.Paths.BinPath }); err != nil {
		return err
	}
	return comp(zip, configs, func(bc crt.BuildConfig) string { return bc.Paths.ZipPath })
})

func comp(name string, bcs *opts.AllBuildConfigs, getPath func(crt.BuildConfig) string) error {
	p, v, err := getSHAs(bcs, getPath)
	if err != nil {
		return err
	}
	if p != v {
		return fmt.Errorf("%s mismatch", name)
	}
	log.Println("OK: %s file reproduced correctly")
	return nil
}

func getSHAs(bcs *opts.AllBuildConfigs, getPath func(crt.BuildConfig) string) (primary, verification string, err error) {
	if primary, err = digest.FileSHA256Hex(getPath(bcs.Primary)); err != nil {
		return
	}
	verification, err = digest.FileSHA256Hex(getPath(bcs.Verification))
	return
}
