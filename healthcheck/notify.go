package healthcheck

import (
	"bytes"
	"context"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/frgrisk/healthchecker/notify"
	"github.com/pkg/errors"
)

func (c Config) Notify(result Result) error {
	// Build template
	var mdPayload bytes.Buffer
	tmpl, err := template.New("notify").Parse(mdTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to parse notification template")
	}
	err = tmpl.Execute(&mdPayload, result)
	if err != nil {
		return errors.Wrap(err, "failed to execute notification template")
	}

	var htmlPayload bytes.Buffer
	tmpl, err = template.New("notify").Parse(plainTemplate)
	if err != nil {
		return errors.Wrap(err, "failed to parse notification template")
	}
	err = tmpl.Execute(&htmlPayload, result)
	if err != nil {
		return errors.Wrap(err, "failed to execute notification template")
	}

	// Teams notification
	if c.TeamsWebhookURL != "" {
		err = notify.Teams(c.TeamsWebhookURL, string(mdPayload.Bytes()))
		if err != nil {
			return errors.Wrap(err, "failed to send notification to Teams webhook")
		}
	}

	// SNS notification
	if len(c.SNSTopicARNs) > 0 {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return errors.Wrap(err, "failed to load AWS SDK configuration")
		}

		client := sns.NewFromConfig(cfg)

		for _, topicARN := range c.SNSTopicARNs {
			input := &sns.PublishInput{
				Message:  aws.String(string(htmlPayload.Bytes())),
				TopicArn: aws.String(topicARN),
			}
			_, err = notify.SNS(context.TODO(), client, input)
			if err != nil {
				return errors.Wrapf(err, "failed to publish SNS message to SNS topic %q", topicARN)
			}
		}
	}
	return nil
}

var mdTemplate = `
# Service {{.Name}} {{.ChangeDescription}}

## Details
- **URL**: {{.URL}}
- **Status**: {{.Status}} ({{.Description}})
- **Response Time**: {{.ResponseTime}}
- **Last Check Time**: {{.LastCheckTime}}
- **Last Successful Check Time**: {{.LastSuccessTime}}
- **Last Failure Time**: {{.LastFailureTime}}
- **Payload**: {{.Body}}
`

var plainTemplate = `
Service {{.Name}} {{.ChangeDescription}}

Details:
	• URL: {{.URL}}
	• Status: {{.Status}} ({{.Description}})
	• Response Time: {{.ResponseTime}}
	• Last Check Time: {{.LastCheckTime}}
	• Last Successful Check Time: {{.LastSuccessTime}}
	• Last Failure Time: {{.LastFailureTime}}
	• Payload: {{.Body}}
`
