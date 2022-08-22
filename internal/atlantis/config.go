package atlantis

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"os"
	"path/filepath"
	"sort"
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

func ConfigToWorkspaces(cfg *valid.RepoCfg) DirectoriesWithWorkspaces {
	workspaces := make(DirectoriesWithWorkspaces)
	for _, p := range cfg.Projects {
		if _, exists := workspaces[p.Dir]; !exists {
			workspaces[p.Dir] = []string{}
		}
		workspaces[p.Dir] = append(workspaces[p.Dir], p.Workspace)
	}
	return workspaces
}

func ParseRepoConfig(body string) (*valid.RepoCfg, error) {
	t := true
	var pv config.ParserValidator
	vg := valid.GlobalCfg{
		Repos: []valid.Repo{
			{
				ID:                   "terraform",
				AllowCustomWorkflows: &t,
			},
		},
	}
	vc, err := pv.ParseRepoCfgData([]byte(body), vg, "terraform")
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %s", err)
	}
	return &vc, nil
}

func ParseRepoConfigFromDir(dir string) (*valid.RepoCfg, error) {
	filename := filepath.Join(dir, config.AtlantisYAMLFilename)
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %s", err)
	}
	return ParseRepoConfig(string(body))
}
