package main

import (
    "time"
)

// Status is a model used to represent the current status of the API server.
type Status struct {
    Name    string      `json:"name"`
    Status  string      `json:"status"`
    Version float32     `json:"version"`
    Time    time.Time   `json:"time"`
}

// Update will update the fields of the Status model it is operating on.
func (o *Status) Update() {
    o.Time = time.Now()
}
