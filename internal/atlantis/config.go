package atlantis

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"gopkg.in/yaml.v3"
)

type DirectoriesWithWorkspaces map[string][]string

func (d DirectoriesWithWorkspaces) SortedKeys() []string {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ConfigToWorkspaces(cfg *SimpleAtlantisConfig) DirectoriesWithWorkspaces {
	workspaces := make(DirectoriesWithWorkspaces)
	for _, p := range cfg.Projects {
		if _, exists := workspaces[p.Dir]; !exists {
			workspaces[p.Dir] = []string{}
		}
		workspaces[p.Dir] = append(workspaces[p.Dir], p.Workspace)
	}
	return workspaces
}

type SimpleAtlantisConfig struct {
	Version  int
	Projects []valid.Project
}

func ParseRepoConfig(body string) (*SimpleAtlantisConfig, error) {
	var ret SimpleAtlantisConfig
	if err := yaml.NewDecoder(strings.NewReader(body)).Decode(&ret); err != nil {
		return nil, fmt.Errorf("error parsing config: %s", err)
	}
	return &ret, nil
}

func ParseRepoConfigFromDir(dir string) (*SimpleAtlantisConfig, error) {
	filename := filepath.Join(dir, "atlantis.yaml")
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %s", err)
	}
	return ParseRepoConfig(string(body))
}
