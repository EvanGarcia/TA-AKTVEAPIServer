package main

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// Match is a struct representing a match between Users of AKTVE.
type Match struct {
	ID           int       `json:"id" bson:"id"`
	Participants []int     `json:"participants" bson:"participants"`
	Messages     []Message `json:"-"` // (NOTE: We don't want to return messages every time a Match struct is returned.)
}

// PullMessages retrieves all of this particular Matches from the server.
func (o *Match) PullMessages() error {
	// Remove all of the Messages from the Match so we can repopulate them from
	// the database (NOTE: This suck. Do this better one day.)
	o.Messages = o.Messages[:0]

	// Switch to the "messages" collection
	c := gDatabase.db.DB(dbDB).C("messages")

	// Retrieve all Messages from this Match
	if err := c.Find(bson.M{"participants": bson.M{"$all": o.Participants}}).All(&o.Messages); err != nil {
		return errors.New("failed to retrieve Messages")
	}

	return nil
}

// PutMessage adds a new Message to the Match and pushes it up to the Database.
func (o *Match) PutMessage(message Message) error {
	// Append the Message to the local Match
	o.Messages = append(o.Messages, message)

	// Switch to the "messages" collection
	c := gDatabase.db.DB(dbDB).C("messages")

	// Push the new Message up to the database
	if err := c.Insert(&message); err != nil {
		return errors.New("failed to push new Message up to database")
	}

	return nil
}

// GetMessage returns the Message with the given ID if the Match has one.
func (o *Match) GetMessage(id int) (Message, error) {
	for _, element := range o.Messages {
		if element.ID == id {
			return element, nil
		}
	}

	return Message{}, errors.New("match: could not find Message with provided ID")
}

// GetMessageIndex returns the index of the Message with the given ID if
// the Match has one.
func (o *Match) GetMessageIndex(id int) (int, error) {
	for index, element := range o.Messages {
		if element.ID == id {
			return index, nil
		}
	}

	return -1, errors.New("match: could not find Message with provided ID")
}
