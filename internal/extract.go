package internal

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	defer func() { _ = r.Close() }()

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

		// Mandatory: everything under modpack/mods/
		if strings.HasPrefix(name, "modpack/mods/") {
			rel := strings.TrimPrefix(name, "modpack/")
			targetPath := filepath.Join(dest, filepath.FromSlash(rel))
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
				continue
			}

			// create parent dirs
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}
			written, copyErr := copyFileContents(rc, targetPath)
			closeErr := rc.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			_ = written
			continue
		}

		// Optionals: modpack/optional/<group>/...
		if strings.HasPrefix(name, "modpack/optional/") {
			rest := strings.TrimPrefix(name, "modpack/optional/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) == 0 || parts[0] == "" {
				continue
			}
			group := parts[0]
			if _, ok := selected[group]; !ok {
				continue // not selected
			}

			// target path should remove the modpack/optional/<group>/ prefix
			rel := ""
			if len(parts) == 2 {
				rel = parts[1]
			} else {
				// directory entry for the group itself
				rel = ""
			}

			if rel == "" {
				// ensure the destination dir exists (no-op)
				if err := os.MkdirAll(dest, 0755); err != nil {
					return err
				}
				continue
			}

			targetPath := filepath.Join(dest, filepath.FromSlash(rel))
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
				continue
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				return err
			}
			written, copyErr := copyFileContents(rc, targetPath)
			closeErr := rc.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			_ = written
			continue
		}

		// ignore other entries
	}

	return nil
}

func copyFileContents(rc io.ReadCloser, targetPath string) (int64, error) {
	out, err := os.Create(targetPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = out.Close() }()

	written, err := io.Copy(out, rc)
	return written, err
}
