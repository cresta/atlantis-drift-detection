package processedcache

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	Client *dynamodb.Client
	Table  string
}

func NewDynamoDB(ctx context.Context, table string) (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	c := DynamoDB{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  table,
	}
	_, err = c.GetRemoteWorkspaces(ctx, &ConsiderWorkspacesChecked{Dir: "test"})
	if err != nil {
		return nil, fmt.Errorf("failed to verify dynamodb cache: %w", err)
	}
	return &c, nil
}

func dynamoKeyForDriftCheckResultKey(keyType string, k fmt.Stringer) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"K": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s:%s", keyType, k.String())},
	}
}

func dynamoKeyForDriftCheckResultValue(keyType string, key fmt.Stringer, value any) (map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	allItems = append(allItems, dynamoKeyForDriftCheckResultKey(keyType, key))
	if i, err := attributevalue.MarshalMap(key); err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	} else {
		allItems = append(allItems, i)
	}
	if i, err := attributevalue.MarshalMap(value); err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	} else {
		allItems = append(allItems, i)
	}
	ret := make(map[string]types.AttributeValue)
	for _, item := range allItems {
		for k, v := range item {
			ret[k] = v
		}
	}
	return ret, nil
}

func (d *DynamoDB) genericGet(ctx context.Context, keyType string, key fmt.Stringer, into any) (bool, error) {
	input := &dynamodb.GetItemInput{
		TableName: &d.Table,
		Key:       dynamoKeyForDriftCheckResultKey(keyType, key),
	}
	output, err := d.Client.GetItem(ctx, input)
	if err != nil {
		return false, fmt.Errorf("failed to get drift check result: %w", err)
	}
	if output.Item == nil {
		return false, nil
	}
	if err := attributevalue.UnmarshalMap(output.Item, into); err != nil {
		return false, fmt.Errorf("failed to unmarshal drift check result: %w", err)
	}
	return true, nil
}

func (d *DynamoDB) genericDelete(ctx context.Context, keyType string, key fmt.Stringer) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &d.Table,
		Key:       dynamoKeyForDriftCheckResultKey(keyType, key),
	}
	_, err := d.Client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete drift check result: %w", err)
	}
	return nil
}

func (d *DynamoDB) genericStore(ctx context.Context, keyType string, key fmt.Stringer, value any) error {
	item, err := dynamoKeyForDriftCheckResultValue(keyType, key, value)
	if err != nil {
		return fmt.Errorf("failed to marshal drift check result: %w", err)
	}
	input := &dynamodb.PutItemInput{
		TableName: &d.Table,
		Item:      item,
	}
	_, err = d.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to store drift check result: %w", err)
	}
	return nil
}

func (d *DynamoDB) GetDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) (*DriftCheckValue, error) {
	var ret DriftCheckValue
	if exists, err := d.genericGet(ctx, "ConsiderDriftChecked", key, &ret); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return &ret, nil
}

func (d *DynamoDB) DeleteDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked) error {
	return d.genericDelete(ctx, "ConsiderDriftChecked", key)
}

func (d *DynamoDB) StoreDriftCheckResult(ctx context.Context, key *ConsiderDriftChecked, value *DriftCheckValue) error {
	return d.genericStore(ctx, "ConsiderDriftChecked", key, value)
}

func (d *DynamoDB) GetRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) (*WorkspacesCheckedValue, error) {
	var ret WorkspacesCheckedValue
	if exists, err := d.genericGet(ctx, "ConsiderWorkspacesChecked", key, &ret); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return &ret, nil
}

func (d *DynamoDB) StoreRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked, value *WorkspacesCheckedValue) error {
	return d.genericStore(ctx, "ConsiderWorkspacesChecked", key, value)
}

func (d *DynamoDB) DeleteRemoteWorkspaces(ctx context.Context, key *ConsiderWorkspacesChecked) error {
	return d.genericDelete(ctx, "ConsiderWorkspacesChecked", key)
}

var _ ProcessedCache = &DynamoDB{}
