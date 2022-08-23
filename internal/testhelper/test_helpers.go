package testhelper

import (
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var once sync.Once

func ReadEnvFile(t *testing.T, rootDir string) {
	once.Do(func() {
		expectedEnvPath := filepath.Join(rootDir, ".env")
		if _, err := os.Stat(expectedEnvPath); err != nil {
			return
		}
		envs, err := godotenv.Read(expectedEnvPath)
		require.NoError(t, err)
		for k, v := range envs {
			if os.Getenv(k) == "" {
				t.Setenv(k, v)
			}
		}
	})
}

func EnvOrSkip(t *testing.T, env string) string {
	body := os.Getenv(env)
	if body == "" {
		t.Skip(env + " not set, skipping test")
	}
	return body
}
