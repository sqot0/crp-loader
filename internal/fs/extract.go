package fs

import (
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractFromZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := f.Name
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimPrefix(name, "/")
		name = filepath.ToSlash(name)

		if f.FileInfo().IsDir() {
			continue
		}

		targetPath := filepath.Join(dest, filepath.FromSlash(name))

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = copyFileContents(rc, targetPath)
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFileContents(rc io.ReadCloser, targetPath string) (int64, error) {
	out, err := os.Create(targetPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	bufWriter := bufio.NewWriter(out)
	defer bufWriter.Flush()

	written, err := io.Copy(bufWriter, rc)
	return written, err
}
