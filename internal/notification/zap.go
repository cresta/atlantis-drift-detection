package notification

import (
	"context"
	"go.uber.org/zap"
)

type Zap struct {
	Logger *zap.Logger
}

func (I *Zap) PlanDrift(_ context.Context, dir string, workspace string) error {
	I.Logger.Info("Plan has drifted", zap.String("dir", dir), zap.String("workspace", workspace))
	return nil
}

func (I *Zap) ExtraWorkspaceInRemote(_ context.Context, dir string, workspace string) error {
	I.Logger.Info("Extra workspace in remote", zap.String("dir", dir), zap.String("workspace", workspace))
	return nil
}

func (I *Zap) MissingWorkspaceInRemote(_ context.Context, dir string, workspace string) error {
	I.Logger.Info("Missing workspace in remote", zap.String("dir", dir), zap.String("workspace", workspace))
	return nil
}

var _ Notification = &Zap{}
