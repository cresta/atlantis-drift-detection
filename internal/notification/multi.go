package notification

import "context"

type Multi struct {
	Notifications []Notification
}

func (m *Multi) TemporaryError(ctx context.Context, dir string, workspace string, err error) error {
	for _, n := range m.Notifications {
		if err := n.TemporaryError(ctx, dir, workspace, err); err != nil {
			return err
		}
	}
	return nil
}

func (m *Multi) ExtraWorkspaceInRemote(ctx context.Context, dir string, workspace string) error {
	for _, n := range m.Notifications {
		if err := n.ExtraWorkspaceInRemote(ctx, dir, workspace); err != nil {
			return err
		}
	}
	return nil
}

func (m *Multi) MissingWorkspaceInRemote(ctx context.Context, dir string, workspace string) error {
	for _, n := range m.Notifications {
		if err := n.MissingWorkspaceInRemote(ctx, dir, workspace); err != nil {
			return err
		}
	}
	return nil
}

func (m *Multi) PlanDrift(ctx context.Context, dir string, workspace string) error {
	for _, n := range m.Notifications {
		if err := n.PlanDrift(ctx, dir, workspace); err != nil {
			return err
		}
	}
	return nil
}

var _ Notification = &Multi{}
