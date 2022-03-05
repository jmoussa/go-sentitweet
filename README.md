# Sentiment Analysis Platform

> A data-acquisition and enrichment pipeline for loading tweets into a MongoDB database with an API on top to query it.
> All wrapped in a neat CLI

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
- Custom (`tw`) CLI utility
    - Run the sentiment analysis pipeline and the API webserver to access sentiment analysis results 
    - *Coming Soon: List, Search, and interact with Twitter through the CLI*


## What the Pipeline Does

the Pipeline streams in tweets (based on a search phrase) and performs sentiment analysis on the tweets, before loading all relevant info into the database.


## Running Locally with the CLI:


```bash
# Project Setup
cd go-sentitweet
cp config/config.json.template config/config.json
# fill in your config.json with your credentials
export CONFIG_LOCATION=$(pwd)/config.json
export PROJECT_ENV=local

cd bin/ # or add to your $PATH
# Run the sentiment analysis pipeline with no tweet search phrases (default #nft)
./tw pipeline 
./tw pipeline --term="#amazon"

# Run the RestAPI server (on port 8080)
./tw server
```