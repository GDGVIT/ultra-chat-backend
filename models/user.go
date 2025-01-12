package models

type User struct {
	ID            string                 `bson:"id"`
	UUID          string                 `bson:"uuid"`
	Token         map[string]interface{} `bson:"token"`
	Username      string                 `bson:"username"`
	Discriminator string                 `bson:"discriminator"`
	Summaries     []Summary              `bson:"summaries,omitempty"`
}

type Summary struct {
	ID      string `bson:"id"`
	Content string `bson:"content"`
}
