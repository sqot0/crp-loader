package fs

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha1.New()

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
