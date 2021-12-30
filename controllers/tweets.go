package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/processors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TweetSearchBody struct {
	SearchPhrase string `json:"searchPhrase"`
}

// POST /tweets
// Get all tweets
func FindTweets(c *gin.Context) {
	var requestBody TweetSearchBody
	if err := c.BindJSON(&requestBody); err != nil {
		log.Fatal(err)
	}
	log.Println("Request: ", requestBody.SearchPhrase)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("twitter-sentiment").Collection("tweets")
	log.Printf("Searching: %s", requestBody.SearchPhrase)
	filterCursor, err := collection.Find(ctx, bson.M{"basetweet.text": bson.M{"$regex": requestBody.SearchPhrase}})
	if err != nil {
		log.Fatal(err)
	}
	defer filterCursor.Close(ctx)

	var tweets []processors.TweetWithScore
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Fatal(err)
	}
	log.Println(len(tweets), " Tweets Found")
	c.JSON(http.StatusOK, gin.H{"data": tweets, "count": len(tweets)})
}

// GET /tweets /:id
// Find a tweet by id
func FindTweet(c *gin.Context) { // Get model if exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("twitter-sentiment").Collection("tweets")

	filterCursor, err := collection.Find(ctx, bson.M{"basetweet.id": c.Param("id")})
	if err != nil {
		log.Fatal(err)
	}
	var tweets []processors.TweetWithScore
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Fatal(err)
	}
	for _, x := range tweets {
		log.Println(x.BaseTweet.Text)
	}
	c.JSON(http.StatusOK, gin.H{"data": tweets})
}
