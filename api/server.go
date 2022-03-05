package api

import (
	"log"
	"net/http"

	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
)

func RunServer() {
	r := gin.Default()
	r.POST("/tweets", FindTweets)
	r.GET("/tweet/:id", FindTweet)
	r.GET("/logs", PipeLogs)
	// use statsviz for program health visualization
	statsviz.RegisterDefault()
	go func() {
		// stat viz for the server is available on port :6060
		log.Println("Navigate to: http://localhost:6060/debug/statsviz/ for metrics")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	r.Run()
}
