package drifter

import (
	"context"
	"fmt"
	"github.com/cresta/atlantis-drift-detection/internal/atlantis"
	"github.com/cresta/atlantis-drift-detection/internal/atlantisgithub"
	"github.com/cresta/atlantis-drift-detection/internal/terraform"
	"github.com/cresta/gogit"
	"github.com/cresta/gogithub"
)

type Drifter struct {
	Repo         string
	Cloner       *gogit.Cloner
	GithubClient gogithub.GitHub
	Terraform    *terraform.Client
	Notification Notification
}

type Notification interface {
	ExtraWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
	MissingWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
}

func (d *Drifter) Drift(ctx context.Context) error {
	repo, err := atlantisgithub.CheckOutTerraformRepo(ctx, d.GithubClient, d.Cloner, d.Repo)
	if err != nil {
		return fmt.Errorf("failed to checkout repo %s: %w", d.Repo, err)
	}
	cfg, err := atlantis.ParseRepoConfigFromDir(repo.Location())
	if err != nil {
		return fmt.Errorf("failed to parse repo config: %w", err)
	}
	workspaces := atlantis.ConfigToWorkspaces(cfg)
	if err := d.FindExtraWorkspaces(ctx, workspaces); err != nil {
		return fmt.Errorf("failed to find extra workspaces: %w", err)
	}
	return nil
}

func (d *Drifter) FindExtraWorkspaces(ctx context.Context, ws atlantis.DirectoriesWithWorkspaces) error {
	for dir, workspaces := range ws {
		if err := d.Terraform.Init(ctx, dir); err != nil {
			return fmt.Errorf("failed to init workspace %s: %w", dir, err)
		}
		var expectedWorkspaces []string
		expectedWorkspaces = append(expectedWorkspaces, workspaces...)
		expectedWorkspaces = append(expectedWorkspaces, "default")
		remoteWorkspaces, err := d.Terraform.ListWorkspaces(ctx, dir)
		if err != nil {
			return fmt.Errorf("failed to list workspaces in %s: %w", dir, err)
		}
		for _, w := range remoteWorkspaces {
			if !contains(expectedWorkspaces, w) {
				if err := d.Notification.ExtraWorkspaceInRemote(ctx, dir, w); err != nil {
					return fmt.Errorf("failed to notify of extra workspace %s in %s: %w", w, dir, err)
				}
			}
		}
	}
	return nil
}

func contains(workspaces []string, w string) bool {
	for _, workspace := range workspaces {
		if workspace == w {
			return true
		}
	}
	return false
}
