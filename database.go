package main

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"
)

// Declare some database settings
// (TODO: Move these to a separate file.)
const dbHost = "db.aktve-app.com" // The hostname of the database
const dbUser = "apiserver"        // The user to connect to the database as
const dbPass = "AktveServerPass"  // The password for the user above
const dbDB = "aktve"

// Database is a light wrapper struct for an MGo MongoDB database session
// object.
type Database struct {
	db *mgo.Session
}

// DatabaseConnect attempts to dial the database and cache a new session with
// it.
func (o *Database) DatabaseConnect() error {
	// If we are connected to the database, disconnected
	if o.DatabaseTest() != nil {
		o.DatabaseDisconnect()
	}

	// Dial the database
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{dbHost},
		Timeout:  60 * time.Second,
		Database: dbDB,
		Username: dbUser,
		Password: dbPass,
	}
	//session, err := mgo.Dial("mongodb://" + dbUser + ":" + dbPass + "@" + dbHost + "/" + dbDB)
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}

	// Cache the new session
	o.db = session

	return nil
}

// DatabaseDisconnect closes the current connection to the database.
func (o *Database) DatabaseDisconnect() {
	// See if we have a session to work with
	if o.db != nil {
		o.db.Close()
	}
}

// DatabaseTest checks to see if we currently have a good connection to the
// database.
func (o *Database) DatabaseTest() error {
	// See if we even think we have a session to work with
	if o.db == nil {
		return errors.New("database: session not dialed")
	}

	// Ping the database to see if it's there
	err := o.db.Ping()

	return err
}

var gDatabase = Database{}
