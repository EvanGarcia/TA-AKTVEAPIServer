// (NOTE: To retrieve query string values, use r.URL.Query().Get("america"). To
// retrieve POST data values, use data.Cheese = r.FormValue("cheese").)

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

// EndpointGETIndex handles the "GET /" API endpoint.
func EndpointGETIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

// EndpointGETStatus handles the "GET /status" API endpoint.
func EndpointGETStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response of the API call
	data := Status{Name: "AKTVE API Server", Status: "online", Version: 1.0}
	data.Update()

	// Create a success response
	success := Success{Success: true, Error: ""}

	// Combine the success and data structs so that they can be returned
	returnJSON := struct {
		Success
		Status
	}{success, data}

	if err := json.NewEncoder(w).Encode(returnJSON); err != nil {
		panic(err)
	}
}

// EndpointPOSTLogin handles the "POST /login" API endpoint.
func EndpointPOSTLogin(w http.ResponseWriter, r *http.Request) {
	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Token string `json:"token,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.FormValue("fb_userid") == "" || r.FormValue("fb_access_token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'fb_userid' and 'fb_access_token' paramaters are required."
	} else {
		// Attempt to get User from the database
		var m bson.M
		c := gDatabase.db.DB(dbDB).C("fb_links")
		if num, err := strconv.Atoi(r.FormValue("fb_userid")); err == nil {
			err := c.Find(bson.M{"fb_user_id": num}).One(&m)
			if err != nil {
				success.Success = false
				success.Error = "Invalid `fb_userid` provided to API call. User does not exist."
			} else {
				// Switch to the sessions collection
				c := gDatabase.db.DB(dbDB).C("sessions")

				// Remove any old sessions for the user
				c.RemoveAll(bson.M{"user_id": m["user_id"]})

				// Generate a new access token (and regenerate it until it is
				// a unique one)
				for {
					data.Token = GenerateToken() // (NOTE: This will get returned.)
					if err := c.Find(bson.M{"token": data.Token}).One(&m); err != nil {
						break
					}
				}

				// Insert the new session into the collection
				c.Insert(&Session{Token: data.Token, UserID: m["user_id"].(int)})
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Are you sure that `fb_userid` is a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeSettings handles the "GET /me/settings" API endpoint.
func EndpointGETMeSettings(w http.ResponseWriter, r *http.Request) {
	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		ShareLocation bool `json:"sharelocation,omitempty"`
		FriendMen     bool `json:"friendmen,omitempty"`
		FriendWomen   bool `json:"friendwomen,omitempty"`
		DateMen       bool `json:"datemen,omitempty"`
		DateWomen     bool `json:"datewomen,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		data.ShareLocation = true
		data.FriendMen = true
		data.FriendWomen = true
		data.DateMen = false
		data.DateWomen = false
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointPOSTMeSettings handles the "POST /me/settings" API endpoint.
func EndpointPOSTMeSettings(w http.ResponseWriter, r *http.Request) {
	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		// (TODO: Post the settings into the database.)
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMe handles the "GET /me" API endpoint.
func EndpointGETMe(w http.ResponseWriter, r *http.Request) {
	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
		Data    User
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data User
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		data = gDemoUsers[0]
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointPUTMe handles the "PUT /me" API endpoint.
func EndpointPUTMe(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars
		// (TODO: Actually save the image somewhere and update the database.)
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointDELETEMe handles the "DELETE /me" API endpoint.
func EndpointDELETEMe(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars
		// (TODO: Actually save the image somewhere and update the database.)
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeMatches handles the "GET /me/matches" API endpoint.
func EndpointGETMeMatches(w http.ResponseWriter, r *http.Request) {
	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Matches []Match `json:"matches"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		data.Matches = gDemoUsers[0].Matches
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeMatchesID handles the "GET /me/matches/{match_id}" API endpoint.
func EndpointGETMeMatchesID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Match Match `json:"match,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if num, err := strconv.Atoi(vars["match_id"]); err == nil {
			if match, err := gDemoUsers[0].GetMatch(num); err == nil {
				data.Match = match
			} else {
				success.Success = false
				success.Error = "Invalid `match_id` provided to API call. Match does not exist for User."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointPOSTMeMatchesIDMessage handles the "POST /me/matches/{match_id}/message" API endpoint.
func EndpointPOSTMeMatchesIDMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else if r.FormValue("message") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'message' paramater must be provided in POST data."
	} else {
		if num, err := strconv.Atoi(vars["match_id"]); err == nil {
			if index, err := gDemoUsers[0].GetMatchIndex(num); err == nil {
				// Create the new message
				message := Message{
					ID:       int(time.Now().Unix() << 32),
					AuthorID: 0,
					Message:  r.FormValue("message"),
					Date:     time.Now().String(),
				}

				// Append it to the list of Messages
				gDemoUsers[0].Matches[index].Messages = append(gDemoUsers[0].Matches[index].Messages, message)
			} else {
				success.Success = false
				success.Error = "Invalid `match_id` provided to API call. Match does not exist for User."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeMatchesIDMessages handles the
// "GET /me/matches/{match_id}/messages" API endpoint.
func EndpointGETMeMatchesIDMessages(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Messages []Message `json:"messages,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if num, err := strconv.Atoi(vars["match_id"]); err == nil {
			if match, err := gDemoUsers[0].GetMatch(num); err == nil {
				data.Messages = match.Messages
			} else {
				success.Success = false
				success.Error = "Invalid `match_id` provided to API call. Match does not exist for User."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeMatchesIDMessagesID handles the
// "GET /me/matches/{match_id}/messages/{message_id}" API endpoint.
func EndpointGETMeMatchesIDMessagesID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Message Message `json:"message,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if num, err := strconv.Atoi(vars["match_id"]); err == nil {
			if match, err := gDemoUsers[0].GetMatch(num); err == nil {
				if num, err := strconv.Atoi(vars["message_id"]); err == nil {
					if message, err := match.GetMessage(num); err == nil {
						data.Message = message
					} else {
						success.Success = false
						success.Error = "Invalid `message_id` provided to API call. Message does not exist for Match."
					}
				} else {
					success.Success = false
					success.Error = "Internal API error. Was the provided `message_id` a valid number?"
				}
			} else {
				success.Success = false
				success.Error = "Invalid `match_id` provided to API call. Match does not exist for User."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETMeMatchesIDMessagesAfterID handles the
// "GET /me/matches/{match_id}/messages/after/{message_id}" API endpoint.
func EndpointGETMeMatchesIDMessagesAfterID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		Messages []Message `json:"messages,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if num, err := strconv.Atoi(vars["match_id"]); err == nil {
			if match, err := gDemoUsers[0].GetMatch(num); err == nil {
				if num, err := strconv.Atoi(vars["message_id"]); err == nil {
					if index, err := match.GetMessageIndex(num); err == nil {
						data.Messages = match.Messages[(index + 1):]
					} else {
						success.Success = false
						success.Error = "Invalid `message_id` provided to API call. Message does not exist for Match."
					}
				} else {
					success.Success = false
					success.Error = "Internal API error. Was the provided `message_id` a valid number?"
				}
			} else {
				success.Success = false
				success.Error = "Invalid `match_id` provided to API call. Match does not exist for User."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointPUTMeImagesID handles the "PUT /me/images/{image_id}" API endpoint.
func EndpointPUTMeImagesID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else if r.FormValue("image_data") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'image_data' paramater must be provided in POST data."
	} else {
		var _ = vars
		// (TODO: Actually save the image somewhere and update the database.)
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETUsersID handles the "GET /user/{user_id}" API endpoint.
func EndpointGETUsersID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		User User `json:"user,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if num, err := strconv.Atoi(vars["user_id"]); err == nil {
			// Attempt to get User from the database
			c := gDatabase.db.DB(dbDB).C("users")
			err = c.Find(bson.M{"id": num}).One(&data.User)
			if err != nil {
				success.Success = false
				success.Error = "Invalid `user_id` provided to API call. User does not exist."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointPUTUsersIDFeeling handles the "PUT /user/{user_id}/feeling" API endpoint.
func EndpointPUTUsersIDFeeling(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type ReturnData struct {
		Success Success
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else if r.FormValue("feeling") != "like" && r.FormValue("feeling") != "dislike" {
		success.Success = false
		success.Error = "Invalid API call. 'feeling' paramater must either be 'like' or 'dislike'."
	} else {
		if num, err := strconv.Atoi(vars["user_id"]); err == nil {
			if user, err := GetUser(num); err == nil {
				var _ = user
				// (TODO: Update the User's feeling and associated matches towards User.)
			} else {
				success.Success = false
				success.Error = "Invalid `user_id` provided to API call. User does not exist."
			}
		} else {
			success.Success = false
			success.Error = "Internal API error. Was the provided `match_id` a valid number?"
		}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}

// EndpointGETPotentials handles the "GET /user/{user_id}" API endpoint.
func EndpointGETPotentials(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Write the HTTP header for the response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Create the actual data response structs of the API call
	type GenericData struct {
		PotentialUserIDs []int `json:"potential_user_ids,omitempty"`
	}

	type ReturnData struct {
		Success Success
		Data    GenericData
	}

	// Create the response structs
	var success Success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if r.URL.Query().Get("token") != "a1b2c3d4e5f6g7h8i9j" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars
		data.PotentialUserIDs = []int{1, 2}
	}

	// Combine the success and data structs so that they can be returned
	returnData.Success = success
	returnData.Data = data

	// Respond with the JSON-encoded return data
	if err := json.NewEncoder(w).Encode(returnData); err != nil {
		panic(err)
	}
}
