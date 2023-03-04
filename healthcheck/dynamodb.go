package healthcheck

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
)

// TableBasics encapsulates the Amazon DynamoDB service actions used in the examples.
// It contains a DynamoDB service client that is used to act on the specified table.
type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func (basics TableBasics) CreateTable() (*types.TableDescription, error) {
	var tableDesc *types.TableDescription

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(partitionKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(partitionKey),
				KeyType:       types.KeyTypeHash,
			},
		},
		TableName:   aws.String(basics.TableName),
		BillingMode: types.BillingModePayPerRequest,
	}

	table, err := basics.DynamoDbClient.CreateTable(context.Background(), input)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create DynamoDB table %v", basics.TableName)
	} else {
		waiter := dynamodb.NewTableExistsWaiter(basics.DynamoDbClient)
		err = waiter.Wait(
			context.TODO(),
			&dynamodb.DescribeTableInput{
				TableName: aws.String(basics.TableName)},
			5*time.Minute,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "wait for table exists for DynamoDB table %v failed", basics.TableName)
		}
		tableDesc = table.TableDescription
	}
	return tableDesc, err
}

// TableExists determines whether a DynamoDB table exists.
func (basics TableBasics) TableExists() (bool, error) {
	exists := true
	_, err := basics.DynamoDbClient.DescribeTable(
		context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(basics.TableName)},
	)
	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			err = nil
		}
		exists = false
	}
	return exists, err
}

func (basics TableBasics) GetItem(url string) (*Result, error) {
	result := &Result{}
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(basics.TableName),
		Key: map[string]types.AttributeValue{
			partitionKey: &types.AttributeValueMemberS{Value: url},
		},
	}
	itemOutput, err := basics.DynamoDbClient.GetItem(context.TODO(), getItemInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last healthcheck from DynamoDB table")
	}
	return result, errors.Wrap(attributevalue.UnmarshalMap(itemOutput.Item, &result), "failed to unmarshal previous healthcheck result")
}

func (basics TableBasics) PutItem(r Result) error {
	resultAV, err := attributevalue.MarshalMap(r)
	if err != nil {
		return errors.Wrap(err, "failed to marshal healthcheck result to DynamoDB item")
	}
	putItemInput := &dynamodb.PutItemInput{
		Item:      resultAV,
		TableName: aws.String(basics.TableName),
	}
	_, err = basics.DynamoDbClient.PutItem(context.TODO(), putItemInput)
	return errors.Wrap(err, "failed to put healthcheck result to DynamoDB table")
}
