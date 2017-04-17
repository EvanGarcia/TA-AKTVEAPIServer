package main

// Success is a model used to represent success or failure of an API call.
type Success struct {
    Success bool    `json:"success"`
    Error   string  `json:"error"`
}
