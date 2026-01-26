package internal

import (
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// ExtractSelectedFromZip extracts only necessary files from the archive to dest.
// Always extracts anything under modpack/mods/ (mandatory), and extracts selected optional groups
// from modpack/optional/<group>/. When placing files, the leading "modpack/" prefix is removed
// so that modpack/mods/foo.jar -> <dest>/mods/foo.jar and modpack/optional/<group>/config -> <dest>/config
func ExtractSelectedFromZip(zipPath, dest string, selectedGroups []string) error {
	selected := map[string]struct{}{}
	for _, s := range selectedGroups {
		selected[s] = struct{}{}
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Pre-scan to collect files to extract and total size for progress
	type fileInfo struct {
		file       *zip.File
		targetPath string
	}
	var filesToExtract []fileInfo
	var totalSize int64
	for _, f := range r.File {
		name := f.Name
		// normalize
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimPrefix(name, "/")
		name = filepath.ToSlash(name)

		// We only care about entries under modpack/
		if !strings.HasPrefix(name, "modpack/") {
			continue
		}

		shouldExtract := false
		rel := ""
		if strings.HasPrefix(name, "modpack/mods/") {
			shouldExtract = true
			rel = strings.TrimPrefix(name, "modpack/")
		} else if strings.HasPrefix(name, "modpack/optional/") {
			rest := strings.TrimPrefix(name, "modpack/optional/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) > 0 && parts[0] != "" {
				if _, ok := selected[parts[0]]; ok {
					shouldExtract = true
					if len(parts) == 2 {
						rel = parts[1]
					}
				}
			}
		}

		if shouldExtract && !f.FileInfo().IsDir() && rel != "" {
			targetPath := filepath.Join(dest, filepath.FromSlash(rel))
			filesToExtract = append(filesToExtract, fileInfo{file: f, targetPath: targetPath})
			totalSize += int64(f.UncompressedSize64)
		}
	}

	// Create progress bar
	bar := progressbar.DefaultBytes(totalSize, "Распаковка")
	var mu sync.Mutex

	// First pass: Create all directories
	for _, fi := range filesToExtract {
		if err := os.MkdirAll(filepath.Dir(fi.targetPath), 0755); err != nil {
			return err
		}
	}

	// Second pass: Extract files with worker pool
	var wg sync.WaitGroup
	errChan := make(chan error, len(filesToExtract))
	taskChan := make(chan fileInfo, len(filesToExtract))

	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fi := range taskChan {
				rc, err := fi.file.Open()
				if err != nil {
					errChan <- err
					continue
				}
				written, copyErr := copyFileContents(rc, fi.targetPath)
				closeErr := rc.Close()
				if copyErr != nil {
					errChan <- copyErr
					continue
				}
				if closeErr != nil {
					errChan <- closeErr
					continue
				}
				mu.Lock()
				bar.Add(int(written))
				mu.Unlock()
				errChan <- nil
			}
		}()
	}

	for _, fi := range filesToExtract {
		taskChan <- fi
	}
	close(taskChan)

	wg.Wait()
	close(errChan)
	for err := range errChan {
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
