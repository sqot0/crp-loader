package manifest

import (
	"encoding/json"
	"net/http"
)

type ModpackInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Optionals   []string `json:"optionals"`
	SHA1        string   `json:"sha1"`
	Priority    int      `json:"priority"`
}

type OptionalInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SHA1        string `json:"sha1"`
}

type Manifest struct {
	Modpacks  map[string]ModpackInfo  `json:"modpacks"`
	Optionals map[string]OptionalInfo `json:"optionals"`
}

func GetManifest(serverURL string) (*Manifest, error) {
	url := serverURL + "manifest.json"
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (m *Manifest) GetModpackByName(name string) (string, *ModpackInfo) {
	for i, mp := range m.Modpacks {
		if mp.Name == name {
			return i, &mp
		}
	}
	return "", nil
}
