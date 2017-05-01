package main

import (
	"github.com/twinj/uuid"
)

// GenerateToken creates a new API access token that is (supposedly) unique and
// can thus be associated with a particular User and their Session.
func GenerateToken() string {
	return uuid.NewV4().String()
}
