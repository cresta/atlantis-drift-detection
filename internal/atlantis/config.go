package atlantis

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"os"
	"path/filepath"
)

type DirectoriesWithWorkspaces map[string][]string

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
	var pv config.ParserValidator
	var vg valid.GlobalCfg
	vc, err := pv.ParseRepoCfgData([]byte(body), vg, "")
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
