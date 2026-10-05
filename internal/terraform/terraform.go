package terraform

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cresta/pipe"
	"go.uber.org/zap"
	"path/filepath"
	"strings"
)

type Client struct {
	Directory string
	Logger    *zap.Logger
	Binary    string
}

func (c *Client) binary() string {
	if c.Binary == "" {
		return "terraform"
	}
	return c.Binary
}

type execErr struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
	root   error
}

func (e *execErr) Unwrap() error {
	return e.root
}

func (e *execErr) Error() string {
	return fmt.Sprintf("%s:%s:%s", e.stdout.String(), e.stderr.String(), e.root.Error())
}

func (c *Client) Init(ctx context.Context, subDir string) error {
	c.Logger.Info("Initializing terraform", zap.String("dir", subDir))
	var stdout, stderr bytes.Buffer
	result := pipe.NewPiped(c.binary(), "init", "-no-color").WithDir(filepath.Join(c.Directory, subDir)).Execute(ctx, nil, &stdout, &stderr)
	if result != nil {
		return &execErr{
			stdout: stdout,
			stderr: stderr,
			root:   result,
		}
	}
	return nil
}

func (c *Client) ListWorkspaces(ctx context.Context, subDir string) ([]string, error) {
	c.Logger.Info("Listing workspaces", zap.String("dir", subDir))
	var stdout, stderr bytes.Buffer
	result := pipe.NewPiped(c.binary(), "workspace", "list").WithDir(filepath.Join(c.Directory, subDir)).Execute(ctx, nil, &stdout, &stderr)
	if result != nil {
		return nil, &execErr{
			stdout: stdout,
			stderr: stderr,
			root:   result,
		}
	}
	lines := strings.Split(stdout.String(), "\n")
	workspaces := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		workspaces = append(workspaces, line)
	}
	return workspaces, nil
}
