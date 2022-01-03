package main

import (
	"log"
	"net/http"

	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/api"
)

func main() {
	r := gin.Default()
	r.POST("/tweets", api.FindTweets)
	r.GET("/tweet/:id", api.FindTweet)
	// use statsviz for program health visualization
	statsviz.RegisterDefault()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	r.Run()
}
