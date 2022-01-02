package processors

import (
	"context"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/grassmudhorses/vader-go/lexicon"
	"github.com/grassmudhorses/vader-go/sentitext"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TweetWithScore struct {
	BaseTweet       *twitter.Tweet
	SentimentScores map[string]float64
}

func LexiconSentimentAnalysis(s interface{}) (interface{}, error) {
	// Takes in an interface{} message, fetchest sentiment scores, and pushes updated message with scores
	tweet := s.(*twitter.Tweet)
	parseText := sentitext.Parse(tweet.Text, lexicon.DefaultLexicon)
	results := sentitext.PolarityScore(parseText)
	//log.Println("Positive:", results.Positive)
	//log.Println("Negative:", results.Negative)
	//log.Println("Neutral:", results.Neutral)
	//log.Println("Compound:", results.Compound)
	var scores map[string]float64
	scores = map[string]float64{
		"Positive": results.Positive,
		"Negative": results.Negative,
		"Neutral":  results.Neutral,
		"Compound": results.Compound,
	}
	var obj TweetWithScore
	obj.BaseTweet = tweet
	obj.SentimentScores = scores
	return obj, nil
}

func FormatAndUpload(s interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("twitter-sentiment").Collection("tweets")
	collection.InsertOne(ctx, s)
	// Takes in a string message, alters it and pushes updated message
	//tweet := s.(TweetWithScore)
	//log.Printf("Process 2: Text - %s\n--------------------------------\n", tweet.BaseTweet.Text)
	return s, nil
}
