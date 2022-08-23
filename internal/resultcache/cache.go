package resultcache

import (
	"context"
	"fmt"
	"time"
)

type DriftCheckResultKey struct {
	// The directory checked
	Dir string
	// The workspace checked
	Workspace string
}

func (d *DriftCheckResultKey) String() string {
	return fmt.Sprintf("%s:%s", d.Dir, d.Workspace)
}

type DriftCheckResultValue struct {
	// If non-empty, indicates an error in the checking
	Error string
	// Only if we have an empty error: the result of checking for drift
	Drift bool `json:"drift"`
	// Only if we have an empty error: when we did this check
	When time.Time
}

type RemoteWorkspacesKey struct {
	// Directory checked
	Dir string
}

func (d *RemoteWorkspacesKey) String() string {
	return d.Dir
}

type RemoteWorkspacesValue struct {
	// If non-empty, indicates an error in the checking
	Error string
	// Worksaces we remember in this remote
	Workspaces []string
	// Only if we have an empty error: when we did this check
	When time.Time
}

type ResultCache interface {
	GetDriftCheckResult(ctx context.Context, key *DriftCheckResultKey) (*DriftCheckResultValue, error)
	DeleteDriftCheckResult(ctx context.Context, key *DriftCheckResultKey) error
	StoreDriftCheckResult(ctx context.Context, key *DriftCheckResultKey, value *DriftCheckResultValue) error
	GetRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey) (*RemoteWorkspacesValue, error)
	StoreRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey, value *RemoteWorkspacesValue) error
	DeleteRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey) error
}

type Noop struct{}

func (n Noop) GetDriftCheckResult(ctx context.Context, key *DriftCheckResultKey) (*DriftCheckResultValue, error) {
	return nil, nil
}

func (n Noop) DeleteDriftCheckResult(ctx context.Context, key *DriftCheckResultKey) error {
	return nil
}

func (n Noop) StoreDriftCheckResult(ctx context.Context, key *DriftCheckResultKey, value *DriftCheckResultValue) error {
	return nil
}

func (n Noop) GetRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey) (*RemoteWorkspacesValue, error) {
	return nil, nil
}

func (n Noop) StoreRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey, value *RemoteWorkspacesValue) error {
	return nil
}

func (n Noop) DeleteRemoteWorkspaces(ctx context.Context, key *RemoteWorkspacesKey) error {
	return nil
}

var _ ResultCache = &Noop{}
