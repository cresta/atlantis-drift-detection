package notification

import (
	"context"
	"sync"

	"github.com/cresta/gogithub"
)

func NewWorkflow(ghClient gogithub.GitHub, owner string, repo string, id string, ref string) *Workflow {
	if owner == "" || repo == "" || id == "" || ref == "" {
		return nil
	}
	return &Workflow{
		WorkflowOwner: owner,
		WorkflowRepo:  repo,
		WorkflowId:    id,
		WorkflowRef:   ref,
		GhClient:      ghClient,
	}
}

type Workflow struct {
	GhClient      gogithub.GitHub
	WorkflowOwner string
	WorkflowRepo  string
	WorkflowId    string
	WorkflowRef   string

	mu              sync.Mutex
	directoriesDone map[string]struct{}
}

func (w *Workflow) TemporaryError(_ context.Context, _ string, _ string, _ error) error {
	// Ignored
	return nil
}

func (w *Workflow) ExtraWorkspaceInRemote(_ context.Context, _ string, _ string) error {
	return nil
}

func (w *Workflow) MissingWorkspaceInRemote(_ context.Context, _ string, _ string) error {
	return nil
}

func (w *Workflow) PlanDrift(ctx context.Context, dir string, _ string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.directoriesDone == nil {
		w.directoriesDone = make(map[string]struct{})
	}
	if _, ok := w.directoriesDone[dir]; ok {
		return nil
	}
	w.directoriesDone[dir] = struct{}{}
	return w.GhClient.TriggerWorkflow(ctx, w.WorkflowOwner, w.WorkflowRepo, w.WorkflowId, w.WorkflowRef, map[string]string{
		"directory": dir,
	})
}

var _ Notification = &Workflow{}
