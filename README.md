# Go Sentiment Analysis

## Components


**Analysis**: Functions available for use in the data pipelines to perform mutations on the data

**API**: handle the API call/logic for fetching tweets and sentiment scores

**CLI**: twitter (tw) CLI utility logic for interacting with tweets and sentiment scores  

**Config**: config module based in JSON (enter twitter credentials for use)

**Data Pipelines**: orchestrate/run tweet crawling and sentiment analysis

**DB**: DB-specific connection and query logic

**Monitoring**: monitoring and logging utilities (using AWS SNS for live-streaming insights through SQS Queue subscriptions)


## Architecture

There are two main pieces of architecture

- API
- Sentiment Analysis/Data Uploader

Additionally there's a monitoring module that will log each message and publish it to an SNS topic for streaming behind the API

To run the API (on port :8080):

```bash
cd go-sentitweet
export CONFIG_LOCATION=$(pwd)/config.json
export PROJECT_ENV=local
go run main.go
```

To run the content fetching and sentiment analysis pipeline:

```bash
cd go-sentitweet
export CONFIG_LOCATION=$(pwd)/config.json
export PROJECT_ENV=local
go run data-pielines/orchestrator.go
```
