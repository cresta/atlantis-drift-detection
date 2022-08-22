package drifter

import (
	"context"
	"fmt"
	"github.com/cresta/atlantis-drift-detection/internal/atlantis"
	"github.com/cresta/atlantis-drift-detection/internal/atlantisgithub"
	"github.com/cresta/atlantis-drift-detection/internal/notification"
	"github.com/cresta/atlantis-drift-detection/internal/terraform"
	"github.com/cresta/gogit"
	"github.com/cresta/gogithub"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Drifter struct {
	Logger             *zap.Logger
	Repo               string
	Cloner             *gogit.Cloner
	GithubClient       gogithub.GitHub
	Terraform          *terraform.Client
	Notification       notification.Notification
	AtlantisClient     *atlantis.Client
	DirectoryWhitelist []string
	SkipWorkspaceCheck bool
	ParallelRuns       int
}

func (d *Drifter) Drift(ctx context.Context) error {
	repo, err := atlantisgithub.CheckOutTerraformRepo(ctx, d.GithubClient, d.Cloner, d.Repo)
	if err != nil {
		return fmt.Errorf("failed to checkout repo %s: %w", d.Repo, err)
	}
	d.Terraform.Directory = repo.Location()
	cfg, err := atlantis.ParseRepoConfigFromDir(repo.Location())
	if err != nil {
		return fmt.Errorf("failed to parse repo config: %w", err)
	}
	workspaces := atlantis.ConfigToWorkspaces(cfg)
	if err := d.FindDriftedWorkspaces(ctx, workspaces); err != nil {
		return fmt.Errorf("failed to find drifted workspaces: %w", err)
	}
	if err := d.FindExtraWorkspaces(ctx, workspaces); err != nil {
		return fmt.Errorf("failed to find extra workspaces: %w", err)
	}
	return nil
}

func (d *Drifter) shouldSkipDirectory(dir string) bool {
	if len(d.DirectoryWhitelist) == 0 {
		return false
	}
	for _, whitelist := range d.DirectoryWhitelist {
		if dir == whitelist {
			return false
		}
	}
	return true
}

type errFunc func(ctx context.Context) error

func (d *Drifter) drainAndExecute(ctx context.Context, toRun []errFunc) error {
	if d.ParallelRuns <= 1 {
		for _, r := range toRun {
			if err := r(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	from := make(chan errFunc)
	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for _, r := range toRun {
			select {
			case from <- r:
			case <-egctx.Done():
				return egctx.Err()
			}
		}
		close(from)
		return nil
	})
	for i := 0; i < d.ParallelRuns; i++ {
		eg.Go(func() error {
			for {
				select {
				case <-egctx.Done():
					return egctx.Err()
				case r, ok := <-from:
					if !ok {
						return nil
					}
					if err := r(egctx); err != nil {
						return err
					}
				}
			}
		})
	}
	return eg.Wait()
}

func (d *Drifter) FindDriftedWorkspaces(ctx context.Context, ws atlantis.DirectoriesWithWorkspaces) error {
	runningFunc := func(dir string) errFunc {
		return func(ctx context.Context) error {
			if d.shouldSkipDirectory(dir) {
				d.Logger.Info("Skipping directory", zap.String("dir", dir))
				return nil
			}
			workspaces := ws[dir]
			d.Logger.Info("Checking for drifted workspaces", zap.String("dir", dir))
			for _, workspace := range workspaces {
				pr, err := d.AtlantisClient.PlanSummary(ctx, &atlantis.PlanSummaryRequest{
					Repo:      d.Repo,
					Ref:       "master",
					Type:      "Github",
					Dir:       dir,
					Workspace: workspace,
				})
				if err != nil {
					return fmt.Errorf("failed to get plan summary for (%s#%s): %w", dir, workspace, err)
				}
				if pr.IsLocked() {
					d.Logger.Info("Plan is locked, skipping drift check", zap.String("dir", dir))
					continue
				}
				if pr.HasChanges() {
					if err := d.Notification.PlanDrift(ctx, dir, workspace); err != nil {
						return fmt.Errorf("failed to notify of plan drift in %s: %w", dir, err)
					}
				}
			}
			return nil
		}
	}
	runs := make([]errFunc, 0)
	for _, dir := range ws.SortedKeys() {
		runs = append(runs, runningFunc(dir))
	}
	return d.drainAndExecute(ctx, runs)
}

func (d *Drifter) FindExtraWorkspaces(ctx context.Context, ws atlantis.DirectoriesWithWorkspaces) error {
	if d.SkipWorkspaceCheck {
		return nil
	}
	runFunc := func(dir string) errFunc {
		return func(ctx context.Context) error {
			if d.shouldSkipDirectory(dir) {
				d.Logger.Info("Skipping directory", zap.String("dir", dir))
				return nil
			}
			workspaces := ws[dir]
			d.Logger.Info("Checking for extra workspaces", zap.String("dir", dir))
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
			return nil
		}
	}
	runs := make([]errFunc, 0)
	for _, dir := range ws.SortedKeys() {
		runs = append(runs, runFunc(dir))
	}
	return d.drainAndExecute(ctx, runs)
}

func contains(workspaces []string, w string) bool {
	for _, workspace := range workspaces {
		if workspace == w {
			return true
		}
	}
	return false
}
