package main

import (
	"context"
	"log"
	"runtime"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/jmoussa/go-sentitweet/config"
	"github.com/jmoussa/go-sentitweet/processors"
	"golang.org/x/sync/semaphore"
)

// Parse JSON config for use
var cfg config.Config = config.ParseConfig()

func producer(ctx context.Context, strings []string) (<-chan interface{}, error) {
	outChannel := make(chan interface{})

	go func() {
		defer close(outChannel)

		for _, s := range strings {
			select {
			case <-ctx.Done():
				return
			case outChannel <- s:
			}
		}
	}()

	return outChannel, nil
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
				if count%1000 == 0 {
					log.Printf("Tweet #: %d", count)
				}
				//tweet := val.(*twitter.Tweet)
				//log.Printf("sink: %s", tweet.Text)
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
		// if cancelled, abort operation
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

		// run a go function to handle processing
		go func(s In) {
			// release the semaphore at the end of this concurrent process
			defer sem1.Release(1)
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

func generator(searchPhrase string, cfg config.Config) chan interface{} {
	out := make(chan interface{})
	go func() {
		con := cfg.General
		c := oauth1.NewConfig(con["consumerkey"], con["consumersecret"])
		token := oauth1.NewToken(con["accesstoken"], con["accesssecret"])
		httpClient := c.Client(oauth1.NoContext, token)
		defer close(out)

		// consume stream messages
		client := twitter.NewClient(httpClient)
		params := &twitter.StreamFilterParams{
			Track:         []string{searchPhrase},
			StallWarnings: twitter.Bool(true),
		}
		stream, err := client.Streams.Filter(params)
		if err != nil {
			log.Fatalln(err)
		}
		defer stream.Stop()
		demux := twitter.NewSwitchDemux()
		demux.Tweet = func(tweet *twitter.Tweet) {
			//log.Printf("%+v\n-----------------------\n", tweet)
			out <- tweet
		}
		for message := range stream.Messages {
			demux.Handle(message)
		}
	}()
	return out
}

func main() {
	// TODO: replace with queue generator function that returns arrays from an external queue
	// source := []string{"FOO", "BAR", "BAX"}
	// generator will continue fetching from the queue until empty then will trigger long poll on the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	/*
		readStream, err := producer(ctx, source)
		if err != nil {
			log.Fatal(err)
		}
	*/
	// use generator as initial producer (outputs a string channel)
	sourceChannel := generator("omicron", cfg)
	outputChannel := make(chan interface{})
	errorChannel := make(chan error)
	//construct pipeline processing stages
	go func() {
		step(ctx, sourceChannel, outputChannel, errorChannel, processors.RunProc1Stage)
	}()
	// next stages are connected to each other
	nextStageChannel := make(chan interface{})
	go func() {
		step(ctx, outputChannel, nextStageChannel, errorChannel, processors.RunProc2Stage)
	}()
	sink(ctx, cancel, nextStageChannel, errorChannel)
}
