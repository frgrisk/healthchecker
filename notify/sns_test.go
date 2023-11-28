package notify

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublishImpl struct{}

func (dt SNSPublishImpl) Publish(ctx context.Context,
	params *sns.PublishInput,
	optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {

	output := &sns.PublishOutput{
		MessageId: aws.String("123"),
	}

	return output, nil
}

type Config struct {
	Message  string `json:"Message"`
	TopicArn string `json:"TopicArn"`
}

var globalConfig Config

func populateConfiguration(t *testing.T) {
	t.Helper()
	globalConfig.Message = "Hello World!"
	globalConfig.TopicArn = "dummyTopic"
}

func TestPublish(t *testing.T) {
	thisTime := time.Now()
	nowString := thisTime.Format("2006-01-02 15:04:05 Monday")
	t.Log("Starting unit test at " + nowString)

	populateConfiguration(t)

	// Build the request with its input parameters
	input := sns.PublishInput{
		Message:  &globalConfig.Message,
		TopicArn: &globalConfig.TopicArn,
	}

	api := &SNSPublishImpl{}

	resp, err := SNS(context.Background(), api, &input)
	if err != nil {
		t.Log("Got an error ...:")
		t.Log(err)
		return
	}

	t.Log("Message ID: " + *resp.MessageId)
}
