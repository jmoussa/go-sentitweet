package models

type Tweet struct {
	ID     uint   `bson:"id" json:"id" gorm:"primary_key"`
	Text   string `bson:"text" json:"text"`
	Author string `bson:"author" json:"author"`
}

type CreateTweetInput struct {
	Text   string `bson:"text" json:"text" binding:"required"`
	Author string `bson:"author" json:"author" binding:"required"`
}
