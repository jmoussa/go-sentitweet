# Go Sentiment Analysis

## Components

**Config**: config module based in JSON (enter twitter credentials for use)
**Controllers**: handle the API db call/logic for fetching tweets + scores
**Data Pipelines**: orchestrate/run sentiment analysis and tweet crawl
**Models**: DB Models (not of much use since transitioning to MongoDB)
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

