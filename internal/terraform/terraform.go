package terraform

import (
	"bytes"
	"context"
	"github.com/cresta/pipe"
	"path/filepath"
	"strings"
)

type Client struct {
	Directory string
}

func (c *Client) Init(ctx context.Context, subDir string) error {
	var stdout, stderr bytes.Buffer
	result := pipe.NewPiped("terraform", "init", "-no-color").WithDir(filepath.Join(c.Directory, subDir)).Execute(ctx, nil, &stdout, &stderr)
	if result != nil {
		return result
	}
	return nil
}

func (c *Client) ListWorkspaces(ctx context.Context, subDir string) ([]string, error) {
	var stdout, stderr bytes.Buffer
	result := pipe.NewPiped("terraform", "workspace", "list").WithDir(filepath.Join(c.Directory, subDir)).Execute(ctx, nil, &stdout, &stderr)
	if result != nil {
		return nil, result
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
