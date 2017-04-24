package main

import (
    "log"
    "net/http"
)

func main() {
    // Connect to the database
    gDatabase.DatabaseConnect()

    // Begin serving and routing API endpoints
    router := NewRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}
