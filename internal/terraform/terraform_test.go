package terraform

import (
	"context"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInit(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	c := Client{
		Directory: testhelper.EnvOrSkip(t, "TERRAFORM_DIR"),
	}
	require.NoError(t, c.Init(context.Background(), testhelper.EnvOrSkip(t, "TERRAFORM_SUBDIR")))
}

func TestInit_Emptydir(t *testing.T) {
	td := t.TempDir()
	c := Client{
		Directory: td,
	}
	require.NoError(t, c.Init(context.Background(), ""))
}
