package main

import (
	"gopkg.in/mgo.v2/bson"
)

// File is a struct representing a File in the database.
type File struct {
	ID     bson.ObjectId `json:"_id" bson:"_id"`
	Data   []byte        `json:"data" bson:"data"`
	Type   string        `json:"type" bson:"type"`
	Length int           `json:"length" bson:"length"`
	MD5    []byte        `json:"md5" bson:"md5"`
}
