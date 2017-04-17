package main

// Message is a struct representing a message between Users in a Match.
type Message struct {
    ID              int         `json:"id"`
    AuthorID        int         `json:"author_id"`
    Message         string      `json:"message"`
    Date            string      `json:"date"`
}
