package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/db"
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
		//log.Fatalf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse HTTP Request Body with error: %s", err)})
		return
	}
	log.Println("Request: ", requestBody.SearchPhrase)
	// init db context context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// init db connection
	client, err := db.OpenMongoClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not disconnect mongo client with error: %s", err)})
		return
	}
	// defer closing of db conn
	defer db.CloseMongoClient(client, ctx)
	// text search
	tweets, err, additional_desc := db.TextSearchQueryMongoClient(client, ctx, requestBody.SearchPhrase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", additional_desc, err)})
		return
	}
	// return
	log.Println(len(tweets), " Tweets Found")
	c.JSON(http.StatusOK, gin.H{"data": tweets, "count": len(tweets)})
}

// GET /tweet/:id
// Find a tweet by id
func FindTweet(c *gin.Context) { // Get model if exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to disconnect mongo client with error: %s", err)})
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
