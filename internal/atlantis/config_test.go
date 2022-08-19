package atlantis

import (
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const exampleAtlantis = `version: 3
projects:
- dir: environments/aws/example
  autoplan:
    when_modified:
    - '*.tf'
- dir: environments/aws/account/datadog
  workspace: dev
  autoplan:
    when_modified:
    - '*.tf'
- dir: environments/aws/account/datadog
  workspace: prod
  autoplan:
    when_modified:
    - '*.tf'
`

func TestParseRepoConfigFromDir(t *testing.T) {
	dirName, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		require.NoError(t, err)
	}(dirName)
	fp := filepath.Join(dirName, config.AtlantisYAMLFilename)
	require.NoError(t, os.WriteFile(fp, []byte(exampleAtlantis), 0644))
	cfg, err := ParseRepoConfigFromDir(dirName)
	require.NoError(t, err)
	require.Equal(t, 3, cfg.Version)
	require.Equal(t, 3, len(cfg.Projects))
	require.Equal(t, "environments/aws/example", cfg.Projects[0].Dir)
}
