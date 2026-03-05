package fs

import (
	"io"
	"net/http"
	"os"
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

	_, err = io.Copy(out, resp.Body)
	return err
}
