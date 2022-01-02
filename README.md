# Go Sentiment Analysis

## Components

**Config**: config module based in JSON (enter twitter credentials for use)

**API**: handle the API call/logic for fetching tweets and sentiment scores

**Data Pipelines**: orchestrate/run tweet crawling and sentiment analysis

**Models**: DB Models (not much use since transitioning to MongoDB)

**Processors**: Functions available for use in the data pipeilines

## Architecture

There are two main pieces of architecture
- API
- Sentiment Analysis/Data Uploader

To run the API (on port :8080):
```bash
go run main.go
```

To run the content fetching and sentiment analysis pipeline:
```bash
go run data-pielines/orchestrator.go
```

