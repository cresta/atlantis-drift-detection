package notification

import (
	"context"
)

type State int

const (
	StateUnknown State = iota
	StateNoDrift
	StateExtraWorkspaceInRemote
	StateMissingWorkspaceInRemote
)

type Location struct {
	Directory string
	Workspace string
}

type Notification interface {
	ExtraWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
	MissingWorkspaceInRemote(ctx context.Context, dir string, workspace string) error
	PlanDrift(ctx context.Context, dir string, workspace string) error
}
