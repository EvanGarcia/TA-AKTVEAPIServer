package main

// Like is a struct representing a Like between from one User of AKTVE of
// another.
type Like struct {
	ID      int `json:"id" bson:"id"`
	LikerID int `json:"liker_id" bson:"liker_id"`
	LikeeID int `json:"likee_id" bson:"likee_id"`
}
