# healthchecker

[![Go Report Card](https://goreportcard.com/badge/github.com/frgrisk/healthchecker)](https://goreportcard.com/report/github.com/frgrisk/healthchecker)

`healthchecker` is a command-line utility for performing health checks on a
specified URL. The utility can be configured to send notifications when the 
HTTP response status code changes or when the URL recovers from a failure.

## Installation

If your system has a [supported version of Go installed](https://go.dev/dl/),
you can build from source:

```bash
go install github.com/frgrisk/healthchecker@latest
```

## Usage

To use the utility, run the `healthchecker run` command followed by the URL to 
monitor and any configuration flags. For example:

```bash
healthchecker run \
    --url https://example.com \
    --timeout 10s \
    --interval 1m \
    --failure-threshold 5 \
    --success-threshold 3 \
    --name example \
    --teams-webhook-url https://example.webhook.office.com/webhookb2/...
```

The `healthchecker run` command takes the following flags:

* `--url`: the URL to monitor (required)
* `--timeout`: the maximum time to wait for a response from the URL (default 1s)
* `--interval`: the interval at which to perform health checks (default 10s)
* `--failure-threshold`: the number of consecutive failures before sending a down notification (default 5)
* `--success-threshold`: the number of consecutive successes before sending a recovered notification (default 3)
* `--name`: the name of the service for notifications (defaults to the URL 
  if not specified)
* `--teams-webhook-url`: the URL of the Microsoft Teams webhook to use for sending notifications
* `--dynamodb-table-name`: the name of the DynamoDB table to use for storing 
  health check results (default "healthchecker-results")
* `--count`: the number of times to run the health check (0 = infinite)
* `--log-filename`: the log file to write to (if not specified, write to stdout)
* `--log-level`: logging level to use 
  ("panic"|"fatal"|"error"|"warning"|"info"|"debug"|"trace") (default "info")

By default, the `healthchecker run` command will run indefinitely, 
performing health checks on the specified URL at the specified interval. If 
the URL returns a status code indicating failure, the utility will send a 
notification to the specified Microsoft Teams webhook. The utility will also 
keep track of the URL's status in a AWS DynamoDB table to avoid sending 
duplicate notifications.

## To-Do

- [x] Add logging redirection
- [ ] Add support for Lambda
- [ ] Add support for other notification providers
- [ ] Add support for other storage providers
- [ ] Add support for other health check types (e.g. TCP, DNS, etc.)
- [ ] Add support for custom headers
- [ ] Add support for custom HTTP methods
- [ ] Add support for custom HTTP bodies
- [ ] Add configuration wizard
- [ ] Add Docker images
