package main

import (
	"net/http"
)

// Route is a model for a route to and endpoint function of the API.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is an array type of route models that can be used to specify hold all
// of the different endpoints of the API.
type Routes []Route

// This is a global array of routes to endpoints for the API. It makes use of
// the Routes type that is exported above, and thus is an array of Route models
// specifying each of the said endpoints. This object is used by the router in
// this project.
var routes = Routes{
	Route{
		"GETIndex",
		"GET",
		"/",
		EndpointGETIndex,
	},
	Route{
		"GETStatus",
		"GET",
		"/status",
		EndpointGETStatus,
	},
	Route{
		"POSTLogin",
		"POST",
		"/login",
		EndpointPOSTLogin,
	},
	Route{
		"GETMeSettings",
		"GET",
		"/me/settings",
		EndpointGETMeSettings,
	},
	Route{
		"POSTMeSettings",
		"POST",
		"/me/settings",
		EndpointPOSTMeSettings,
	},
	Route{
		"GETMe",
		"GET",
		"/me",
		EndpointGETMe,
	},
	Route{
		"PUTMe",
		"PUT",
		"/me",
		EndpointPUTMe,
	},
	Route{
		"DELETEMe",
		"DELETE",
		"/me",
		EndpointDELETEMe,
	},
	Route{
		"GETMeMatches",
		"GET",
		"/me/matches",
		EndpointGETMeMatches,
	},
	Route{
		"GETMeMatchesID",
		"GET",
		"/me/matches/{match_id}",
		EndpointGETMeMatchesID,
	},
	Route{
		"POSTMeMatchesIDMessage",
		"POST",
		"/me/matches/{match_id}/message",
		EndpointPOSTMeMatchesIDMessage,
	},
	Route{
		"GETMeMatchesIDMessages",
		"GET",
		"/me/matches/{match_id}/messages",
		EndpointGETMeMatchesIDMessages,
	},
	Route{
		"GETMeMatchesIDMessagesID",
		"GET",
		"/me/matches/{match_id}/messages/{message_id}",
		EndpointGETMeMatchesIDMessagesID,
	},
	Route{
		"GETMeMatchesIDMessagesAfterID",
		"GET",
		"/me/matches/{match_id}/messages/after/{message_id}",
		EndpointGETMeMatchesIDMessagesAfterID,
	},
	Route{
		"PUTMeImagesID",
		"PUT",
		"/me/images/{image_id}",
		EndpointPUTMeImagesID,
	},
	Route{
		"GETUsersID",
		"GET",
		"/users/{user_id}",
		EndpointGETUsersID,
	},
	Route{
		"PUTUsersIDFeeling",
		"PUT",
		"/users/{user_id}/feeling",
		EndpointPUTUsersIDFeeling,
	},
	Route{
		"GETPotentials",
		"GET",
		"/potentials",
		EndpointGETPotentials,
	},
}
