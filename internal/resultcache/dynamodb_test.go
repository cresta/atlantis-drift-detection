package resultcache

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeTestClient(t *testing.T) *DynamoDB {
	testhelper.ReadEnvFile(t, "../../")
	dynamoTable := testhelper.EnvOrSkip(t, "DYNAMODB_TABLE")
	cfg, err := config.LoadDefaultConfig(context.Background())
	require.NoError(t, err)
	c := DynamoDB{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  dynamoTable,
	}
	return &c
}

func TestDynamoDB(t *testing.T) {
	GenericCacheWorkflowTest(t, makeTestClient(t))
}
