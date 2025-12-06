package internal

import (
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads url to filepathDst (overwrites if exists)
func DownloadFile(url, filepathDst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := os.Create(filepathDst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}
