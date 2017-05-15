package main

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// Session is a struct representing a User of AKTVE's Session.
type Session struct {
	Token  string `json:"token" bson:"token"`
	UserID int    `json:"user_id" bson:"user_id"`
}

// SessionCache is a struct used for locally caching Sessions so that the
// database doesn't have to constantly be accessed
type SessionCache struct {
	Sessions []Session
}

var gSessionCache SessionCache

// CheckSession attempts to verify a Session with the provided Token. If the
// Session verifies, the associated User's ID is returned.
func (o *SessionCache) CheckSession(token string) (int, error) {
	// Start by looking for the Session in the cache
	for _, element := range o.Sessions {
		if element.Token == token {
			return element.UserID, nil
		}
	}

	// If the Session is not in the cache, check the database for it (and cache
	// it if it is found)
	var session Session

	c := gDatabase.db.DB(dbDB).C("sessions")
	if err := c.Find(bson.M{"token": token}).One(&session); err != nil {
		return -1, errors.New("could not find Session with provided Token")
	}

	gSessionCache.Sessions = append(gSessionCache.Sessions, session)

	return session.UserID, nil
}

// CreateSession creates a new Session (and associated token) for the User with
// the provided ID, removes any old Sessions for said User, adds the new
// Session to the database and cache, and finally returns the new Session.
func (o *SessionCache) CreateSession(userID int) (Session, error) {
	// Remove any old sessions for the user
	o.CleanSessions(userID)

	// Switch to the sessions database collection
	c := gDatabase.db.DB(dbDB).C("sessions")

	// Generate a new access token (and regenerate it until it is
	// a unique one)
	var session = Session{UserID: userID}
	var m bson.M
	for {
		session.Token = GenerateToken()
		if err := c.Find(bson.M{"token": session.Token}).One(&m); err != nil {
			break
		}
	}

	// Insert the new Session into the database collection
	c.Insert(session)

	// Add the new Session to the cache
	gSessionCache.Sessions = append(gSessionCache.Sessions, session)

	return session, nil
}

// CleanSessions removes all sessions associated with the User with the
// provided ID.
func (o *SessionCache) CleanSessions(userID int) error {
	// Switch to the sessions database collection
	c := gDatabase.db.DB(dbDB).C("sessions")

	// Remove any old Sessions for the User from the database
	c.RemoveAll(bson.M{"user_id": userID})

	// Remove any old Sessions for the User from the cache
	for index, element := range o.Sessions {
		if element.UserID == userID {
			gSessionCache.Sessions = append(gSessionCache.Sessions[:index], gSessionCache.Sessions[(index+1):]...)
		}
	}

	return nil
}
