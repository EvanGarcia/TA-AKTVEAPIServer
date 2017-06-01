package main

import (
	"errors"
	"math"

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

// IsMatchedWith returns whether the User is currently matched with the User
// with the provided ID.
func (o *User) IsMatchedWith(userID int) bool {
	for _, element := range o.Matches {
		for value := range element.Participants {
			if value == userID {
				return true
			}
		}
	}

	return false
}

// CurrentlyLikes returns whether the User currently already likes the User with
// the provided ID.
func (o *User) CurrentlyLikes(userID int) bool {
	c := gDatabase.db.DB(dbDB).C("likes")
	if cnt, err := c.Find(bson.M{"liker_id": o.ID, "likee_id": userID}).Count(); err == nil && cnt > 0 {
		return true
	}

	return false
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

// PullMatches updates the local User object with all of the actual Matches
// between the given User and other Users.
func (o *User) PullMatches() error {
	var likes []Like
	var like Like

	// Switch to the "likes" collection
	c := gDatabase.db.DB(dbDB).C("likes")

	// Retrieve all likes of the User
	if err := c.Find(bson.M{"likee_id": o.ID}).All(&likes); err != nil {
		return errors.New("failed to retrieve Matches")
	}

	// Clear out the local cache of the User's Matches so that we can rebuild
	// it (NOTE: This sucks and is terribly inefficient. One day, if we aren't
	// lazy or exhausted, we should change how all of this works.)
	o.Matches = o.Matches[:0]

	// Figure out which ones are from Users that the User likes back
	for _, element := range likes {
		if err := c.Find(bson.M{"liker_id": o.ID, "likee_id": element.LikerID}).One(&like); err != nil {
			return errors.New("could not retrieve Match")
		}

		// Construct a new Match for the located likes and append it to the
		// User's slice of Matches
		o.Matches = append(o.Matches, Match{ID: len(o.Matches), Participants: []int{o.ID, element.LikerID}})
	}

	return nil
}

// DistanceFrom returns the distance that the User is from the specified
// position. (TODO: This is a dumb algorithm calculated based on a flat plane.
// The actual app currently uses an algorithm that takes the curvature of the
// earth into account. We should switch this function to use that algorithm.)
func (o *User) DistanceFrom(latitude float32, longitude float32) float32 {
	x := math.Pow((float64)(latitude-o.Latitude), 2)
	y := math.Pow((float64)(longitude-o.Longitude), 2)

	return (float32)(math.Sqrt(x + y))
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

// DeleteUser literally deletes the User with the specified ID from both the
// local cache and the database. It should be used for account deletion.
func (o *UserCache) DeleteUser(userID int) error {
	// Delete User from local cache
	for index, element := range gUserCache.Users {
		if element.ID == userID {
			o.Users = append(o.Users[:index], o.Users[(index+1):]...)
		}
	}

	// Delete User from database
	c := gDatabase.db.DB(dbDB).C("users")
	if err := c.Remove(bson.M{"id": userID}); err != nil {
		return errors.New("failed to remove user from database")
	}

	// Delete all of the User's social media links from database
	c = gDatabase.db.DB(dbDB).C("fb_links")
	if err := c.Remove(bson.M{"user_id": userID}); err != nil {
		return errors.New("failed to remove user's Facebook links from database")
	}

	// Delete all of the User's sessions from database
	c = gDatabase.db.DB(dbDB).C("sessions")
	if err := c.Remove(bson.M{"user_id": userID}); err != nil {
		return errors.New("failed to remove user's sessions from database")
	}

	return nil
}
