package internal

import (
	"archive/zip"
	"sort"
	"strings"
)

// InspectOptionalGroups scans the zip and returns the names of directories directly under modpack/optional/
func InspectOptionalGroups(zipPath string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	set := map[string]struct{}{}
	for _, f := range r.File {
		name := f.Name
		// normalize
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimPrefix(name, "/")
		if strings.Contains(name, "modpack/optional/") {
			rest := strings.SplitN(strings.SplitAfter(name, "modpack/optional/")[1], "/", 2)[0]
			if rest != "" {
				set[rest] = struct{}{}
			}
		}
	}

	res := make([]string, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	sort.Strings(res)
	return res, nil
}
