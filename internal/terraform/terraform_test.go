package terraform

import (
	"context"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/cresta/pipe"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"path/filepath"
	"testing"
)

func TestClient_Init(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	c := Client{
		Directory: testhelper.EnvOrSkip(t, "TERRAFORM_DIR"),
		Logger:    zaptest.NewLogger(t),
	}
	require.NoError(t, c.Init(context.Background(), testhelper.EnvOrSkip(t, "TERRAFORM_SUBDIR")))
}

func TestClient_InitEmptydir(t *testing.T) {
	td := t.TempDir()
	c := Client{
		Directory: td,
		Logger:    zaptest.NewLogger(t),
	}
	require.NoError(t, c.Init(context.Background(), ""))
}

func TestClient_ListWorkspaces(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	td := t.TempDir()
	c := Client{
		Directory: td,
		Logger:    zaptest.NewLogger(t),
	}
	const subdir = ""
	require.NoError(t, c.Init(context.Background(), subdir))
	workspaces, err := c.ListWorkspaces(context.Background(), subdir)
	require.NoError(t, err)
	require.Equal(t, []string{"default"}, workspaces)
	ctx := context.Background()
	require.NoError(t, pipe.NewPiped("terraform", "workspace", "new", "testing").WithDir(filepath.Join(c.Directory)).Run(ctx))
	workspaces, err = c.ListWorkspaces(context.Background(), subdir)
	require.NoError(t, err)
	require.Equal(t, []string{"default", "testing"}, workspaces)
}
