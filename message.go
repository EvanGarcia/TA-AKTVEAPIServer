package main

// Message is a struct representing a message between Users in a Match.
type Message struct {
	ID           int    `json:"id" bson:"id"`
	AuthorID     int    `json:"author_id" bson:"author_id"`
	Message      string `json:"message" bson:"message"`
	Date         string `json:"date" bson:"date"`
	Participants []int  `json:"-" bson:"participants"`
}
