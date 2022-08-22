package notification

import (
	"context"
)

type Notification interface {
	ExtraWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
	MissingWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
	PlanDrift(ctx context.Context, dir string, workspace string) error
}
