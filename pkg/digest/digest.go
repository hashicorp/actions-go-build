package digest

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

// FileSHA256Hex calculates the SHA256 sum of file and returns it as a
// hex string.
func FileSHA256Hex(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
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
