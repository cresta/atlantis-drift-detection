package notification

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func genericNotificationTest(t *testing.T, notification Notification) {
	ctx := context.Background()
	require.NoError(t, notification.ExtraWorkspaceInRemote(ctx, "genericNotificationTest/ExtraWorkspaceInRemote", "test-workspace"))
	require.NoError(t, notification.MissingWorkspaceInRemote(ctx, "genericNotificationTest/MissingWorkspaceInRemote", "test-workspace"))
	require.NoError(t, notification.PlanDrift(ctx, "genericNotificationTest/PlanDrift", "test-workspace"))
}
