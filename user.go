package main

import (
    "errors"
)

// User is a struct representing a User of AKTVE.
type User struct {
    ID			int				`json:"id"`
    Name		string			`json:"name"`
    Age			int				`json:"age"`
    Interests	map[string]int	`json:"interests"`
    Tags		[]string		`json:"tags"`
    Bio			string			`json:"bio"`
    Images		[]string		`json:"images"`
    Matches     []Match         `json:"-"`  // (NOTE: We don't want to return matches every time a user struct is returned.)
    Latitude	float32			`json:"latitude"`
    Longitude	float32			`json:"longitude"`
    LastActive	string			`json:"last_active"`
}

// GetMatch returns the Match with the given ID if the User has one.
func (o *User) GetMatch(id int) (Match, error) {
    for _, element := range o.Matches {
        if (element.ID == id) {
            return element, nil
        }
    }

    return Match {}, errors.New("user: could not find Match with provided ID")
}

// GetMatchIndex returns the index of the Match with the given ID if the
// User has one.
func (o *User) GetMatchIndex(id int) (int, error) {
    for index, element := range o.Matches {
        if (element.ID == id) {
            return index, nil
        }
    }

    return -1, errors.New("user: could not find Match with provided ID")
}

var gDemoUsers = []User {
    User {
        ID: 0,
        Name: "Samson",
        Age: 26,
        Interests: map[string]int { "Swimming": 2, "Running": 5 },
        Tags: []string{ "friends_men", "friends_women" },
        Bio: "I'm a pretty cool person.",
        Images: []string { "img/samples/sam1.jpg", "img/samples/sam2.jpg", "img/samples/sam3.jpg" },
        Matches: []Match {
            Match { ID: 0, Participants: []int { 0, 1 }, Messages: []Message { Message { ID: 4003, AuthorID: 0, Message: "Hey man!", Date: "2017-04-17 13:13:09.084228157 -0700 PDT" }, Message { ID: 4004, AuthorID: 1, Message: "Whatup?", Date: "2017-04-17 14:13:09.084228157 -0700 PDT" } } },
            Match { ID: 1, Participants: []int { 0, 2 }, Messages: []Message { Message { ID: 4035, AuthorID: 2, Message: "Yo!", Date: "2017-04-17 13:13:09.084228157 -0700 PDT" }, Message { ID: 4046, AuthorID: 0, Message: "Want to go hiking?", Date: "2017-04-17 14:13:09.084228157 -0700 PDT" } } },
        },
        Latitude: 47.6062,
        Longitude: -122.3321,
        LastActive: "2017-02-15T11:47:26.371Z",
    },
    User {
        ID: 1,
        Name: "Titus",
        Age: 24,
        Interests: map[string]int { "Hiking": 2, "Lifting": 5 },
        Tags: []string{ "friends_men", "friends_women", "dates_women" },
        Bio: "I'm an even cooler person.",
        Images: []string { "img/samples/evan1.jpg" },
        Matches: []Match { },
        Latitude: 47.6062,
        Longitude: -122.3321,
        LastActive: "2017-01-15T11:47:26.371Z",
    },
    User {
        ID: 2,
        Name: "Annie",
        Age: 22,
        Interests: map[string]int { "Snorkeling": 2, "Climbing": 5 },
        Tags: []string{ "friends_men", "friends_women", "dates_men" },
        Bio: "My hair is blond.",
        Images: []string { "http://lorempixel.com/output/people-q-c-100-100-9.jpg" },
        Matches: []Match { },
        Latitude: 47.6062,
        Longitude: -122.3321,
        LastActive: "2017-01-15T11:47:26.371Z",
    },
}

// GetUser returns the User with the given ID if they exist.
func GetUser(id int) (User, error) {
    for _, element := range gDemoUsers {
        if (element.ID == id) {
            return element, nil
        }
    }

    return User {}, errors.New("user: could not find User with provided ID")
}
