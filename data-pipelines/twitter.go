package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/arl/statsviz"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/jmoussa/go-sentitweet/analysis"
	"github.com/jmoussa/go-sentitweet/config"
	"github.com/jmoussa/go-sentitweet/monitoring"
	"golang.org/x/sync/semaphore"
)

/*
Sentiment Analysis Pipeline
Starts multiple goroutines to run the sentiment analysis pipeline concurrently on the number of cores available
*/

// Parse JSON config for use
var cfg config.Config = config.ParseConfig()

func generator(searchPhrase string, cfg config.Config) chan interface{} {
	// Starts up a generator stream of tweets into the outputted channel
	out := make(chan interface{})
	go func() {
		defer close(out)

		con := cfg.General
		c := oauth1.NewConfig(con["consumerkey"], con["consumersecret"])
		token := oauth1.NewToken(con["accesstoken"], con["accesssecret"])
		httpClient := c.Client(oauth1.NoContext, token)

		// intialize stream
		client := twitter.NewClient(httpClient)
		params := &twitter.StreamFilterParams{
			Track:         []string{searchPhrase},
			StallWarnings: twitter.Bool(true),
		}
		stream, err := client.Streams.Filter(params)
		if err != nil {
			log.Fatalf("Error querying stream, %s\n", err)
		}
		defer stream.Stop()

		// Initialize demux for interface{} type processing to channel
		demux := twitter.NewSwitchDemux()
		log.Println("Searching for:", searchPhrase)
		demux.Tweet = func(tweet *twitter.Tweet) {
			out <- tweet
		}
		for message := range stream.Messages {
			demux.Handle(message)
		}
	}()
	return out
}

func mergeAtomic(outputChan chan interface{}, cs ...<-chan interface{}) <-chan interface{} {
	// Atomically dump each channel into the output channel and return output channel
	var i int32
	atomic.StoreInt32(&i, int32(len(cs)))
	for _, c := range cs {
		go func(c <-chan interface{}) {
			for v := range c {
				outputChan <- v
			}
			if atomic.AddInt32(&i, -1) == 0 {
				close(outputChan)
			}
		}(c)
	}
	return outputChan
}

func sink(ctx context.Context, cancelFunc context.CancelFunc, values <-chan interface{}, errors <-chan error) {
	var count int64 = 0
	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err().Error())
			return
		case err := <-errors:
			if err != nil {
				log.Println("error: ", err.Error())
				cancelFunc()
			}
		case _, ok := <-values:
			if ok {
				count += 1
				if count%100 == 0 {
					log.Printf("Tweet count: %d", count)
				}
			} else {
				log.Print("done")
				return
			}
		}
	}
}

func step[In any, Out any](
	ctx context.Context,
	inputChannel <-chan In,
	outputChannel chan Out,
	errorChannel chan error,
	fn func(In) (Out, error),
	loggingTrace string,
) {
	defer close(outputChannel)

	// create a new semaphore with a limit (of the CPU count) for processes the semaphore can access at a time
	limit := runtime.NumCPU()
	sem1 := semaphore.NewWeighted(int64(limit))

	// parse through messages in input channel
	for s := range inputChannel {
		select {
		// if cancelled, abort operation otherwise run while there's values in inputChannel
		case <-ctx.Done():
			log.Println("1 abort")
			break
		default:
		}
		// use semaphores to keep data integrity
		if err := sem1.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
			break
		}

		// start up go functions to parallelize processing to CPU Count
		go func(s In) {
			// release the semaphore at the end of this concurrent process
			defer sem1.Release(1)
			msg, err := json.Marshal(s)
			if err != nil {
				log.Println("Error marshalling: ", err)
			}
			msgStr := string(msg)
			// send and schedule start and stop log messages to SNS
			log := monitoring.Log{
				Message:   msgStr,
				Level:     "INFO",
				Type:      "Start",
				Timestamp: monitoring.GetTimestamp(),
			}
			monitoring.SendLogMessageToSNS(&log)
			log.Type = "Stop"
			defer monitoring.SendLogMessageToSNS(&log)

			// Take the result of the function and send to outputChannel
			result, err := fn(s)
			if err != nil {
				errorChannel <- err
			} else {
				outputChannel <- result
			}
		}(s)
	}

	// after everything's finished fetch and lock the semaphore
	if err := sem1.Acquire(ctx, int64(limit)); err != nil {
		log.Printf("Failed to acquire semaphore: %v", err)
	}
}

func main() {
	statsviz.RegisterDefault()

	// extract search phrase from command line arguments
	var defaultSearchPhrase string
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		defaultSearchPhrase = "#nft"
		log.Println("No search phrase provided, using default:", defaultSearchPhrase)
	} else {
		defaultSearchPhrase = argsWithoutProg[0]
		log.Println("Searching Twitter for:", defaultSearchPhrase)
	}

	go func() {
		log.Println("Navigate to: http://localhost:6070/debug/statsviz/ for metrics")
		log.Println(http.ListenAndServe("localhost:6070", nil))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	/*
		readStream, err := producer(ctx, source)
		if err != nil {
			log.Fatal(err)
		}
	*/
	// using generator as initial producer (outputs an interface{} channel)
	sourceChannel := generator(defaultSearchPhrase, cfg)
	errorChannel := make(chan error)
	// Run lexicon sentiment analysis concurrently with ML Sentiment Analysis
	// then merge the results with the original document?

	// Layer 1: Sentiment Analysis
	layer1OutputChannel := make(chan interface{})
	go func() {
		step(ctx, sourceChannel, layer1OutputChannel, errorChannel, analysis.LexiconSentimentAnalysis, "lexiconSentimentAnalysis")
	}()

	// Layer 2: DB Upload
	layer3OutputChannel := make(chan interface{})
	go func() {
		step(ctx, layer1OutputChannel, layer3OutputChannel, errorChannel, analysis.FormatAndUpload, "formatAndUpload")
	}()

	// Sink
	sink(ctx, cancel, layer3OutputChannel, errorChannel)
}
