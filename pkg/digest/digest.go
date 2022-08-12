package digest

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func SHA256Hex(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SHA256HexStrings(s ...string) (string, error) {
	return SHA256Hex(bytes.NewBufferString(strings.Join(s, "")))
}

func JSONSHA256Hex(a any) (string, error) {
	buf := &bytes.Buffer{}
	mw := io.MultiWriter(buf, os.Stdout)
	if err := json.NewEncoder(mw).Encode(a); err != nil {
		return "", err
	}
	return SHA256Hex(buf)
}

// FileSHA256Hex calculates the SHA256 sum of file and returns it as a
// hex string.
func FileSHA256Hex(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	var closeErr error
	defer func() { closeErr = f.Close() }()
	s, err := SHA256Hex(f)
	if err != nil {
		return s, err
	}
	return s, closeErr
}

func FilesSHA256Hex(names ...string) ([]string, error) {
	out := make([]string, len(names))
	for i, f := range names {
		var err error
		if out[i], err = FileSHA256Hex(f); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func Equal(names ...string) (bool, error) {
	if len(names) < 2 {
		return false, fmt.Errorf("must supply at least 2 files to compare")
	}
	sums, err := FilesSHA256Hex(names...)
	if err != nil {
		return false, err
	}
	last := sums[0]
	for _, s := range sums[1:] {
		if s != last {
			return false, nil
		}
		last = s
	}
	return true, nil
}
