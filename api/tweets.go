package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/config"
	"github.com/jmoussa/go-sentitweet/db"
	"github.com/jmoussa/go-sentitweet/processors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type TweetSearchBody struct {
	SearchPhrase string `json:"searchPhrase,omitempty"`
	DaysBack     int    `json:"daysBack,omitempty"`
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
	var (
		tweets          []processors.TweetWithScoreMessage
		additional_desc string
	)
	if requestBody.SearchPhrase != "" {
		tweets, err, additional_desc = db.TextSearchQueryMongoClient(client, ctx, requestBody.SearchPhrase)
	} else if requestBody.DaysBack > 0 {
		tweets, err, additional_desc = db.FetchRecentTweets(client, ctx, requestBody.DaysBack)
	}
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
	var cfg config.Config = config.ParseConfig()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.General["mongo_url_string"]))
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
	var tweets []processors.TweetWithScoreMessage
	if err = filterCursor.All(ctx, &tweets); err != nil {
		log.Fatal(err)
	}
	for _, x := range tweets {
		log.Println(x.BaseTweet.Text)
	}
	c.JSON(http.StatusOK, gin.H{"data": tweets})
}

// GET /logs
// Fetch Logs
func PipeLogs(c *gin.Context) {
	// Connect to sqs and send messages as they come in
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)
	queueURL := "https://sqs.us-east-1.amazonaws.com/462366532346/logging.fifo"
	waitTime := int64(20)
	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   &waitTime,
	})
	if err != nil {
		log.Fatal(err)
	}
	data := make([]interface{}, 0)
	for _, message := range result.Messages {
		// remove excapes
		raw_message, err := json.Marshal(*message.Body)
		if err != nil {
			log.Fatal(err)
		}
		var logjson interface{}
		err = json.Unmarshal([]byte(raw_message), &logjson)
		if err != nil {
			log.Fatal(err)
		}
		// append to return slice
		data = append(data, logjson)
	}
	fmt.Printf("Retrieved %d messages\n", len(data))
	c.JSON(http.StatusOK, gin.H{"data": data})
}
