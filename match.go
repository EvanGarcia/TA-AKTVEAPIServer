package main

import (
	"errors"
)

// Match is a struct representing a match between Users of AKTVE.
type Match struct {
	ID           int       `json:"id" bson:"id"`
	Participants []int     `json:"participants" bson:"participants"`
	Messages     []Message `json:"-"` // (NOTE: We don't want to return messages every time a Match struct is returned.)
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
