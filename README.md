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
- Sentiment Analysis Pipelines
- *(tw) Twitter CLI coming soon*
    - *List, Search, and interact with Twitter through the CLI*
    - *Access Sentiment Analysis Results

## Running Locally:
```bash
cd go-sentitweet
cp config/config.json.template config/config.json
# fill in your config.json with your credentials
export CONFIG_LOCATION=$(pwd)/config.json
export PROJECT_ENV=local

# To run the sentiment analysis pipeline (will start a local instance that scrapes tweets and sends them to the queue)
# TODO: include CLI command/flag to run the pipeline 
go run data-pielines/twitter.go

# To run the API (on port 8080)
# TODO: include CLI command/flag to run the api 
go run main.go
```
