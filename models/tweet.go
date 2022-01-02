package models

type CreateTweetInput struct {
	Text   string `bson:"text" json:"text" binding:"required"`
	Author string `bson:"author" json:"author" binding:"required"`
}
