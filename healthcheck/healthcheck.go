package healthcheck

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
)

const partitionKey = "url"

type Config struct {
	Name              string
	URL               string
	DynamoDBTableName string
	FailureThreshold  int
	SuccessThreshold  int
	Timeout           time.Duration
	Interval          time.Duration
	SNSTopicARNs      []string
	TeamsWebhookURL   string
	LogFilename       string
	LogLevel          string
}

type Result struct {
	Name                 string    `dynamodbav:"name"`
	URL                  string    `dynamodbav:"url"`
	Body                 string    `dynamodbav:"body"`
	Status               int       `dynamodbav:"status"`
	Description          string    `dynamodbav:"description"`
	ChangeDescription    string    `dynamodbav:"change_description"`
	LastNotificationTime time.Time `dynamodbav:"last_notification_time"`
	LastCheckTime        time.Time `dynamodbav:"last_check_time"`
	ResponseTime         string    `dynamodbav:"response_time"`
	LastSuccessTime      time.Time `dynamodbav:"last_successful_check_time"`
	LastFailureTime      time.Time `dynamodbav:"last_failure_time"`
	SuccessCount         int       `dynamodbav:"success_count"`
	FailureCount         int       `dynamodbav:"failure_count"`
}

func (c Config) Run() (result Result, err error) {
	result.URL = c.URL
	if c.Name == "" {
		c.Name = c.URL
	}
	result.Name = c.Name
	// Create a HTTP client
	client := http.Client{
		Timeout: c.Timeout,
	}
	// Load the previous health check result from DynamoDB
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return result, errors.Wrap(err, "failed to load AWS SDK configuration")
	}
	svc := dynamodb.NewFromConfig(cfg)
	basics := TableBasics{
		DynamoDbClient: svc,
		TableName:      c.DynamoDBTableName,
	}
	exists, err := basics.TableExists()
	if !exists {
		_, err = basics.CreateTable()
		if err != nil {
			return result, errors.Wrap(err, "failed to create DynamoDB table")
		}
	}
	prevResult, err := basics.GetItem(c.URL)
	if err != nil {
		return result, errors.Wrap(err, "failed to get previous healthcheck result from DynamoDB")
	}

	// Send the HTTP request
	req, err := http.NewRequest(http.MethodGet, c.URL, nil)
	if err != nil {
		return result, errors.Wrap(err, "failed to create HTTP request")
	}
	result.LastCheckTime = time.Now()
	res, err := client.Do(req)
	result.ResponseTime = time.Since(result.LastCheckTime).String()
	if err != nil {
		result.Status = http.StatusServiceUnavailable
		result.Description = err.Error()
	} else {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
		// Check the HTTP response
		result.Status = res.StatusCode
		result.Description = http.StatusText(res.StatusCode)
		resBytes, _ := io.ReadAll(res.Body)
		result.Body = string(resBytes)
	}

	// Update the health check status in DynamoDB
	shouldNotify := false
	if result.Status == http.StatusOK {
		result.LastSuccessTime = result.LastCheckTime
		result.LastFailureTime = prevResult.LastFailureTime
		result.SuccessCount = prevResult.SuccessCount + 1
		if result.SuccessCount == c.SuccessThreshold+1 {
			// reset failure count
			result.FailureCount = 0
			shouldNotify = true
		}
	} else {
		// reset success count on any failure
		result.SuccessCount = 0
		result.LastFailureTime = result.LastCheckTime
		result.LastSuccessTime = prevResult.LastSuccessTime
		result.FailureCount = prevResult.FailureCount + 1
		if result.FailureCount == c.FailureThreshold+1 {
			shouldNotify = true
		}
	}
	if shouldNotify {
		result.LastNotificationTime = result.LastCheckTime
		if result.Status == http.StatusOK {
			result.ChangeDescription = "has recovered"
		} else {
			result.ChangeDescription = "is down"
		}
		err = c.Notify(result)
		if err != nil {
			return result, errors.Wrap(err, "failed to send notification(s)")
		}
	}
	return result, errors.Wrap(basics.PutItem(result), "failed to write healthcheck result to DynamoDB")
}
