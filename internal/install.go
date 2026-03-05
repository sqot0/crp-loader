package internal

import (
	"fmt"
	"os"
	"path"

	"github.com/sqot0/crp-loader/internal/fs"
)

type InstallStatus int

const (
	StatusWaiting InstallStatus = iota
	StatusDownloading
	StatusExtracting
	StatusFinished
	StatusError
)

type ProgressUpdate struct {
	ID     string
	Status InstallStatus
	Error  error
}

func Install(id string, downloadURL string, sha1Hash string, progressChan chan ProgressUpdate) {
	sendUpdate := func(status InstallStatus, err error) {
		progressChan <- ProgressUpdate{ID: id, Status: status, Error: err}
	}

	executable, err := os.Executable()
	if err != nil {
		sendUpdate(StatusError, err)
		return
	}
	mcDir := path.Dir(executable)

	filename := path.Base(downloadURL)

	crpLoaderDir := path.Join(mcDir, ".crp-loader")
	if err := os.MkdirAll(crpLoaderDir, os.ModePerm); err != nil {
		sendUpdate(StatusError, err)
		return
	}

	localFile := path.Join(crpLoaderDir, filename)

	_, err = os.Stat(localFile)
	isFileExist := err == nil

	if isFileExist {
		fileHash, err := fs.HashFile(localFile)
		if err == nil && fileHash == sha1Hash {
			sendUpdate(StatusExtracting, nil)
			if err := fs.ExtractFromZip(localFile, mcDir); err != nil {
				sendUpdate(StatusError, err)
				return
			}
			sendUpdate(StatusFinished, nil)
			return
		}
	}

	if err := os.RemoveAll(localFile); err != nil {
		// ignore
	}

	sendUpdate(StatusDownloading, nil)
	if err := fs.DownloadFile(downloadURL, localFile); err != nil {
		sendUpdate(StatusError, err)
		return
	}

	fileHash, err := fs.HashFile(localFile)
	if err != nil {
		sendUpdate(StatusError, err)
		return
	}
	if fileHash != sha1Hash {
		sendUpdate(StatusError, fmt.Errorf("hash mismatch for %s", filename))
		return
	}

	sendUpdate(StatusExtracting, nil)
	if err := fs.ExtractFromZip(localFile, mcDir); err != nil {
		sendUpdate(StatusError, err)
		return
	}

	sendUpdate(StatusFinished, nil)
}
