package db

import (
	"context"
	"log"

	"github.com/jmoussa/go-sentitweet/processors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func OpenMongoClient(ctx context.Context) (*mongo.Client, error) {
	// MongoDB Connection
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
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

func TextSearchQueryMongoClient(client *mongo.Client, ctx context.Context, searchPhrase string) ([]processors.TweetWithScore, error, string) {
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
	var tweets []processors.TweetWithScore
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Print(err)
		return nil, err, "failed to Search DB"
	}
	return tweets, nil, ""
}
