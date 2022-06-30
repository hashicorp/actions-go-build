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
	f, err := os.Open("file.txt")
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
