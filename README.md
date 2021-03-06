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

The pipeline streams in tweets (based on a search phrase) and performs sentiment analysis on the tweets, before loading all relevant info into the database.

**The Process**:

- Fetch Tweets (based on search term)
- Score Tweets (using the `vader-go` Default Lexicon) 
- Upload to DB (MongoDB)

Credentials are configured using JSON config file.

## Running Locally with the CLI:


```bash
# Project Setup (inside project directory)
cp config/config.json.template config/config.json
# fill in your config.json with your credentials
export CONFIG_LOCATION=$(pwd)/config.json

cd bin/ # or add to your $PATH
# Run the sentiment analysis pipeline with no tweet search phrases (default #nft)
# runs in the foreground
./tw pipeline 
./tw pipeline --term="#amazon"

# Run the RestAPI server (on port 8080)
./tw server

# (Coming Soon) Output CSV of tweets and sentiment scores
# running the pipeline chunking 100 tweets at a time to csv
tw pipeline --output=csv --chunk-size=100 --term="#amazon" --output-path=./output/
# query the API/database for tweets by time back
tw query --days_back=5 --output=csv --output-path=./output/
```