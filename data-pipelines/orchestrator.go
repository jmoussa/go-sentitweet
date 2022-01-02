package main

import (
	"context"
	"log"
	"runtime"
	"sync/atomic"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/jmoussa/go-sentitweet/config"
	"github.com/jmoussa/go-sentitweet/processors"
	"golang.org/x/sync/semaphore"
)

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
			log.Print("1 abort")
			break
		default:
		}
		// use semaphores to keep data integrity
		if err := sem1.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
			break
		}

		// start up go functions to parallelize processing to CPU Cound
		go func(s In) {
			// release the semaphore at the end of this concurrent process
			defer sem1.Release(1)
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	/*
		readStream, err := producer(ctx, source)
		if err != nil {
			log.Fatal(err)
		}
	*/
	// using generator as initial producer (outputs an interface{} channel)
	sourceChannel := generator("covid", cfg)
	errorChannel := make(chan error)
	// Run lexicon sentiment analysis concurrently with ML Sentiment Analysis
	// then merge the results with the original document?

	// Layer 1: Sentiment Analysis
	layer1OutputChannel := make(chan interface{})
	go func() {
		step(ctx, sourceChannel, layer1OutputChannel, errorChannel, processors.LexiconSentimentAnalysis)
	}()

	// Layer 2: DB Upload
	layer3OutputChannel := make(chan interface{})
	go func() {
		step(ctx, layer1OutputChannel, layer3OutputChannel, errorChannel, processors.FormatAndUpload)
	}()

	// Sink
	sink(ctx, cancel, layer3OutputChannel, errorChannel)
}
