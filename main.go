package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoussa/go-sentitweet/controllers"
)

func main() {
	r := gin.Default()
	r.POST("/tweets", controllers.FindTweets)
	r.GET("/tweets/:id", controllers.FindTweet)
	r.Run()
}
