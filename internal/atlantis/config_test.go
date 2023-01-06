package atlantis

import (
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

const exampleFromGithubIssue = `version: 3
automerge: true
delete_source_branch_on_merge: true
parallel_plan: true
parallel_apply: true
allowed_regexp_prefixes:
- lab/
- staging/
- prod/
projects:
- name: pepe-ue2-lab-cloudtrail
  workspace: pepe-ue2-lab
  workflow: workflow-1
  dir: components/terraform/cloudtrail
  terraform_version: v1.2.9
  delete_source_branch_on_merge: false
  autoplan:
    enabled: true
    when_modified:
    - '**/*.tf'
    - $PROJECT_NAME.tfvars.json
  apply_requirements:
  - approved`

func TestParseRepoConfig(t *testing.T) {
	_, err := ParseRepoConfig(exampleFromGithubIssue)
	require.NoError(t, err)
}

func TestParseRepoConfigFromDir(t *testing.T) {
	dirName, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		require.NoError(t, err)
	}(dirName)
	fp := filepath.Join(dirName, "atlantis.yaml")
	require.NoError(t, os.WriteFile(fp, []byte(exampleAtlantis), 0644))
	cfg, err := ParseRepoConfigFromDir(dirName)
	require.NoError(t, err)
	require.Equal(t, 3, cfg.Version)
	require.Equal(t, 3, len(cfg.Projects))
	require.Equal(t, "environments/aws/example", cfg.Projects[0].Dir)
}
