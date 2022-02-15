package db

import (
	"context"
	"log"
	"time"

	"github.com/jmoussa/go-sentitweet/config"
	"github.com/jmoussa/go-sentitweet/processors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func OpenMongoClient(ctx context.Context) (*mongo.Client, error) {
	// MongoDB Connection
	var cfg config.Config = config.ParseConfig()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.General["mongo_url_string"]))
	if err != nil {
		return client, err
	}
	return client, nil
}

func CloseMongoClient(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Error: failed to disconnect mongo client with error: %s", err)
		return
		//panic(err)
	}
}

func FetchRecentTweets(client *mongo.Client, ctx context.Context, daysBack int) ([]processors.TweetWithScoreMessage, error, string) {
	// Fetch tweets that have createdat after daysBack
	collection := client.Database("twitter-sentiment").Collection("tweets")
	now := time.Now()
	date := now.AddDate(0, 0, -daysBack)
	log.Printf("**DB Searching: %d days back after %s", daysBack, date.Format("Mon Jan 2 15:04:05 -0700 2006"))
	searchParam := bson.M{"basetweet.createdat": bson.M{"$lte": date.Format("Mon Jan 2 15:04:05 -0700 2006")}}
	// Acquire Query Cursor
	findOptions := options.Find()
	// -1 sorts descending
	findOptions.SetSort(bson.D{{"basetweet.createdat", -1}})
	filterCursor, err := collection.Find(ctx, searchParam, findOptions)
	if err != nil {
		log.Print(err)
		return nil, err, "failed to apply DB Query"
	}
	defer filterCursor.Close(ctx)
	// Run search w/filter
	var tweets []processors.TweetWithScoreMessage
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Print(err)
		return nil, err, "failed to Search DB"
	}
	return tweets, nil, ""

}

func TextSearchQueryMongoClient(client *mongo.Client, ctx context.Context, searchPhrase string) ([]processors.TweetWithScoreMessage, error, string) {
	// MongoDB Query
	collection := client.Database("twitter-sentiment").Collection("tweets")
	log.Printf("Searching: %s", searchPhrase)
	searchParam := bson.M{}
	if len(searchPhrase) > 0 {
		searchParam = bson.M{"basetweet.text": bson.M{"$regex": searchPhrase}}
	}
	// Acquire Query Cursor
	filterCursor, err := collection.Find(ctx, searchParam)
	if err != nil {
		log.Print(err)
		return nil, err, "failed to apply DB Query"
	}
	defer filterCursor.Close(ctx)
	// Run search w/filter
	var tweets []processors.TweetWithScoreMessage
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Print(err)
		return nil, err, "failed to Search DB"
	}
	return tweets, nil, ""
}
