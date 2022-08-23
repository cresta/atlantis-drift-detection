package processedcache

import (
	"context"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeTestClient(t *testing.T) *DynamoDB {
	testhelper.ReadEnvFile(t, "../../")
	client, err := NewDynamoDB(context.Background(), testhelper.EnvOrSkip(t, "DYNAMODB_TABLE"))
	require.NoError(t, err)
	return client
}

func TestDynamoDB(t *testing.T) {
	GenericCacheWorkflowTest(t, makeTestClient(t))
}
