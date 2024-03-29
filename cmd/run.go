package cmd

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/frgrisk/healthchecker/healthcheck"
	nanoid "github.com/matoous/go-nanoid/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Fixed nanoid parameters used
const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	length   = 12
)

func newNanoID() string { return nanoid.MustGenerate(alphabet, length) }

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run performs health checks on the specified URL",
	Run: func(cmd *cobra.Command, args []string) {
		logLevel, err := log.ParseLevel(config.LogLevel)
		if err != nil {
			log.Infof(`Invalid log level %q specified. Using default "info" level.`, config.LogLevel)
			logLevel = log.InfoLevel
		}
		log.SetLevel(logLevel)
		if config.LogFilename != "" {
			file, err := os.OpenFile(config.LogFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				log.Info("Logging to file: ", config.LogFilename)
				log.SetOutput(file)
			} else {
				log.Error("Failed to log to file, using default stderr")
			}
		}
		if config.Interval < config.Timeout {
			log.Fatal("interval must be greater than or equal to timeout")
		}
		var exit bool
		interval := time.NewTicker(config.Interval)
		start := true
		for !exit {
			if start {
				start = false
				performHealthCheck(&count)
			}
			select {
			case <-interval.C:
				if count == 0 {
					exit = true
					break
				}
				performHealthCheck(&count)
			}
		}
		interval.Stop()
	},
}

func performHealthCheck(count *int) {
	logFields := log.Fields{
		"request_id": newNanoID(),
		"url":        config.URL,
	}
	*count = *count - 1
	log.WithFields(logFields).Info("running health check")
	result, err := config.Run()
	if err != nil {
		log.WithFields(logFields).Error(err)
	} else {
		if result.Status != http.StatusOK {
			log.WithFields(logFields).Warn("health check failed")
		} else {
			log.WithFields(logFields).Info("health check succeeded")
		}
	}
}

var config healthcheck.Config

var count int

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&config.Name, "name", "", "name of the service for notifications")
	runCmd.Flags().StringVar(&config.URL, "url", "", "URL of the web service to monitor")
	runCmd.Flags().StringVar(&config.DynamoDBTableName, "dynamodb-table-name", "healthchecker-results", "name of the DynamoDB table to use for storing health check results")
	runCmd.Flags().IntVar(&config.FailureThreshold, "failure-threshold", 5, "number of consecutive failures before sending a down notification")
	runCmd.Flags().IntVar(&config.SuccessThreshold, "success-threshold", 3, "number of consecutive successes before sending a recovered notification")
	runCmd.Flags().DurationVar(&config.Interval, "interval", 10*time.Second, "interval between health checks")
	runCmd.Flags().DurationVar(&config.Timeout, "timeout", time.Second, "timeout for the health check")
	runCmd.Flags().IntVar(&count, "count", 0, "number of times to run the health check (0 = infinite)")
	runCmd.Flags().StringSliceVar(&config.SNSTopicARNs, "sns-topic-arns", []string{}, "ARNs of the SNS topics to use for sending notifications")
	runCmd.Flags().StringVar(&config.TeamsWebhookURL, "teams-webhook-url", "", "URL of the Microsoft Teams webhook to use for sending notifications")
	runCmd.Flags().StringVar(&config.LogFilename, "log-filename", "", "filename to use for logging (default to stdout)")
	runCmd.Flags().StringVar(&config.LogLevel, "log-level", "info", "logging level to use "+printValidFlagValues(allLogLevels()))
	_ = runCmd.MarkFlagRequired("url")
	_ = runCmd.MarkFlagFilename("logging-filename")
}

func allLogLevels() []string {
	var levels []string
	for _, level := range log.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

func printValidFlagValues(i []string) string {
	o := "("
	for _, v := range i {
		o += strconv.Quote(v)
		if v != i[len(i)-1] {
			o += "|"
		}
	}
	o = o + ")"
	return o
}
