package terraform

import (
	"context"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/cresta/pipe"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"os"
	"path/filepath"
	"strings"
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

func TestClient_Binary(t *testing.T) {
	td := t.TempDir()
	argsFile := filepath.Join(td, "args")
	binary := filepath.Join(td, "fake-terraform")
	script := "#!/bin/sh\necho \"$@\" >> " + argsFile + "\necho default\n"
	require.NoError(t, os.WriteFile(binary, []byte(script), 0o755))
	c := Client{
		Directory: td,
		Logger:    zaptest.NewLogger(t),
		Binary:    binary,
	}
	ctx := context.Background()
	require.NoError(t, c.Init(ctx, ""))
	workspaces, err := c.ListWorkspaces(ctx, "")
	require.NoError(t, err)
	require.Equal(t, []string{"default"}, workspaces)
	calls, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.Equal(t, []string{"init -no-color", "workspace list"}, strings.Split(strings.TrimSpace(string(calls)), "\n"))
}
