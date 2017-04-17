package main

import (
    "net/http"
    
    "github.com/gorilla/mux"
)

// NewRouter creates a new HTTP router using Gorilla MUX for each route
// specified in the global "routes" array. See routes.go for the declaration of
// said array.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}
