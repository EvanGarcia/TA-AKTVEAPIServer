package main

import (
	"log"
	"net/http"
	"time"
)

// Logger spins off a new endpoint handler, as specified, and then keeps track
// of and outputs details about its execution.
// TODO: This would be a good place to introduce concurrency.
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
