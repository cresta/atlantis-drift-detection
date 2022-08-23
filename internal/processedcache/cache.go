package processedcache

import (
	"context"
	"fmt"
	"time"
)

type ConsiderDriftChecked struct {
	// The directory checked
	Dir string
	// The workspace checked
	Workspace string
}

func (d *ConsiderDriftChecked) String() string {
	return fmt.Sprintf("%s:%s", d.Dir, d.Workspace)
}

type DriftCheckValue struct {
	// If non-empty, indicates an error in the checking
	Error string
	// Only if we have an empty error: the result of checking for drift
	Drift bool `json:"drift"`
	// Only if we have an empty error: when we did this check
	When time.Time
}

type ConsiderWorkspacesChecked struct {
	// Directory checked
	Dir string
}

func (d *ConsiderWorkspacesChecked) String() string {
	return d.Dir
}

type WorkspacesCheckedValue struct {
	// If non-empty, indicates an error in the checking
	Error string
	// Worksaces we remember in this remote
	Workspaces []string
	// Only if we have an empty error: when we did this check
	When time.Time
}

type ProcessedCache interface {
	GetDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) (*DriftCheckValue, error)
	DeleteDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) error
	StoreDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked, value *DriftCheckValue) error
	GetRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) (*WorkspacesCheckedValue, error)
	StoreRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked, value *WorkspacesCheckedValue) error
	DeleteRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) error
}

type Noop struct{}

func (n Noop) GetDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) (*DriftCheckValue, error) {
	return nil, nil
}

func (n Noop) DeleteDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) error {
	return nil
}

func (n Noop) StoreDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked, value *DriftCheckValue) error {
	return nil
}

func (n Noop) GetRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) (*WorkspacesCheckedValue, error) {
	return nil, nil
}

func (n Noop) StoreRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked, value *WorkspacesCheckedValue) error {
	return nil
}

func (n Noop) DeleteRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) error {
	return nil
}

var _ ProcessedCache = &Noop{}
