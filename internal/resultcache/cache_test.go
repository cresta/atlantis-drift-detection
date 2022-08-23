package resultcache

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func GenericCacheWorkflowTest(t *testing.T, cache ResultCache) {
	currentTime := time.Now().Round(time.Millisecond)
	testKey := &DriftCheckResultKey{
		Dir:       "test" + currentTime.String(),
		Workspace: "test",
	}
	testValue := &DriftCheckResultValue{
		Error: "test",
		Drift: true,
		When:  currentTime,
	}
	ctx := context.Background()
	item, err := cache.GetDriftCheckResult(ctx, testKey)
	require.NoError(t, err)
	require.Nil(t, item)
	err = cache.StoreDriftCheckResult(ctx, testKey, testValue)
	require.NoError(t, err)
	item, err = cache.GetDriftCheckResult(ctx, testKey)
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, testValue, item)
	err = cache.DeleteDriftCheckResult(ctx, testKey)
	require.NoError(t, err)
	item, err = cache.GetDriftCheckResult(ctx, testKey)
	require.NoError(t, err)
	require.Nil(t, item)
}
