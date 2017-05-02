package main

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// User is a struct representing a User of AKTVE.
type User struct {
	ID            int            `json:"id,omitempty" bson:"id"`
	Name          string         `json:"name,omitempty" bson:"name"`
	Age           int            `json:"age,omitempty" bson:"age"`
	Interests     map[string]int `json:"interests,omitempty" bson:"interests"`
	Tags          []string       `json:"tags,omitempty" bson:"tags"`
	Bio           string         `json:"bio,omitempty" bson:"bio"`
	Images        []string       `json:"images,omitempty" bson:"images"`
	Matches       []Match        `json:"-" bson:"-"` // (NOTE: We don't want to return matches every time a user struct is returned.)
	Latitude      float32        `json:"latitude,omitempty" bson:"latitude"`
	Longitude     float32        `json:"longitude,omitempty" bson:"longitude"`
	LastActive    string         `json:"last_active,omitempty" bson:"last_active"`
	ShareLocation bool           `json:"share_location,omitempty" bson:"share_location"`
}

// UserCache is a local cache of User objects used to decrease the number of
// back-and-forth trips between the database server and the API server.
type UserCache struct {
	Users []User
}

var gUserCache UserCache

// GetMatch returns the Match with the given ID if the User has one.
func (o *User) GetMatch(id int) (Match, error) {
	for _, element := range o.Matches {
		if element.ID == id {
			return element, nil
		}
	}

	return Match{}, errors.New("could not find Match with provided ID")
}

// GetMatchIndex returns the index of the Match with the given ID if the
// User has one.
func (o *User) GetMatchIndex(id int) (int, error) {
	for index, element := range o.Matches {
		if element.ID == id {
			return index, nil
		}
	}

	return -1, errors.New("could not find Match with provided ID")
}

// Push updates the User object in the database with its current local
// representation.
func (o *User) Push() error {
	// Update the User in the database
	c := gDatabase.db.DB(dbDB).C("users")
	query := bson.M{"id": o.ID}
	change := bson.M{"$set": o}
	err := c.Update(query, change)

	return err
}

// GetUser retrieves a copy of the User with the specified ID, along with their
// associated UserCache index. If the User is not currently in the UserCache,
// they are retrieved from the database and put into it. If no User is found,
// an error is returned.
func (o *UserCache) GetUser(userID int) (User, int, error) {
	// Check the cache first to see if we already have a local copy of the User
	for index, element := range gUserCache.Users {
		if element.ID == userID {
			return element, index, nil
		}
	}

	// If not in the cache, check the database
	var user User

	c := gDatabase.db.DB(dbDB).C("users")
	if err := c.Find(bson.M{"id": userID}).One(&user); err != nil {
		return User{}, -1, errors.New("could not find User with provided ID")
	}

	gUserCache.Users = append(gUserCache.Users, user)

	return user, (len(gUserCache.Users) - 1), nil
}
