package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/api"
)

func main() {
	r := gin.Default()
	r.POST("/tweets", api.FindTweets)
	r.GET("/tweet/:id", api.FindTweet)
	r.Run()
}
