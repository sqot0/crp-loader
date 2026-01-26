package internal

import (
	"io"
	"net/http"
	"os"

	"github.com/schollz/progressbar/v3"
)

// DownloadFile downloads url to filepathDst (overwrites if exists)
func DownloadFile(url, filepathDst string) error {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepathDst)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Установка")

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}
