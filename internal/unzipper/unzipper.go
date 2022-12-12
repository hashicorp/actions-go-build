// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package unzipper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/actions-go-build/internal/log"
)

type Unzipper struct {
	log log.Func
}

func New(logFunc log.Func) *Unzipper {
	return &Unzipper{log: logFunc}
}

func (uz *Unzipper) Unzip(file, dest string) error {
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = r.Close() }()
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		// Prevent directory traversal.
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path %q", target)
		}
		if err := uz.unzipFile(target, f); err != nil {
			return err
		}
	}
	return closeErr
}

func (uz *Unzipper) unzipFile(target string, f *zip.File) error {
	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, f.Mode())
	}

	uz.log("Extracting file: %s", target)
	if err := os.MkdirAll(filepath.Dir(target), f.Mode()); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = rc.Close() }()
	if err := uz.writeFile(target, rc); err != nil {
		return err
	}
	return closeErr
}

func (uz *Unzipper) writeFile(target string, r io.Reader) error {
	t, err := os.Create(target)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() { closeErr = t.Close() }()
	if _, err := io.Copy(t, r); err != nil {
		return err
	}
	return closeErr
}
