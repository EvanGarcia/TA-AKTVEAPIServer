// (NOTE: To retrieve query string values, use r.URL.Query().Get("america"). To
// retrieve POST data values, use data.Cheese = r.FormValue("cheese").)

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	fb "github.com/huandu/facebook"
	"gopkg.in/mgo.v2/bson"
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
	data := Status{Name: "AKTVE API Server", Status: "online", Version: gAPIVersion}
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
	fbAccessToken := r.FormValue("fb_access_token")

	if fbAccessToken == "" {
		success.Success = false
		success.Error = "Invalid API call. 'fb_access_token' paramater is required."
	} else {
		var m bson.M

		// Try to get the User from Facebook
		if res, err := fb.Get(("/me"), fb.Params{
			"fields":       []string{"id", "name", "picture.width(640)"},
			"access_token": r.FormValue("fb_access_token"),
		}); err != nil {
			success.Success = false
			success.Error = "Invalid `fb_access_token` provided to API call."
		} else {
			// Get the scoped Facebook User ID provided by Facebook itself
			fbUserID := -1
			if str, ok := res["id"].(string); ok {
				if val, err := strconv.Atoi(str); err == nil {
					fbUserID = val
				}
			}
			if fbUserID == -1 {
				success.Success = false
				success.Error = "Could not retrieve scoped Facebook User ID from Facebook."
			} else {
				// Attempt to get User from the database
				c := gDatabase.db.DB(dbDB).C("fb_links")
				if err := c.Find(bson.M{"fb_user_id": fbUserID}).One(&m); err != nil { // If the User is not linked in our database
					// Calculate age based on birthday
					// (NOTE: This is rough...Facebook can only be gauranteed
					// to give us the year.)
					// (NOTE: Skip this for now. Getting a user's birthday from
					// Facebook requires app review.)
					age := 18
					/*if str, ok := res["birthday"].(string); ok {
						slashIndex := strings.Index(str, "/")
						if slashIndex > -1 {
							str = str[(slashIndex + 1):]
						}

						birthdayNumber, _ := strconv.Atoi(str)
						age = (time.Now().Year() - birthdayNumber)
					}*/

					// Figure out what this User's ID will be
					c := gDatabase.db.DB(dbDB).C("users")
					count, _ := c.Count()

					// Get name
					name := ""
					if str, ok := res["name"].(string); ok {
						name = str
					}

					// Get profile picture URL
					profilePictures := []string{}
					if pictureObject, ok := res["picture"].(map[string]interface{}); ok {
						if dataObject, ok := pictureObject["data"].(map[string]interface{}); ok {
							if str, ok := dataObject["url"].(string); ok {
								profilePictures = append(profilePictures, str)
							}
						}
					}

					// Create the new User object
					user := User{
						ID:            count,
						Name:          name,
						Age:           age,
						Interests:     map[string]int{},
						Tags:          []string{},
						Bio:           "",
						Images:        profilePictures,
						Matches:       []Match{},
						Latitude:      0,
						Longitude:     0,
						LastActive:    time.Now().String(),
						ShareLocation: true,
					}

					// Insert the User into the database
					c.Insert(user)

					// Insert the Facebook link for the user into the database
					c = gDatabase.db.DB(dbDB).C("fb_links")
					c.Insert(bson.M{"user_id": user.ID, "fb_user_id": fbUserID, "fb_access_token": r.FormValue("fb_access_token")})

					// Create the new Session for the user and return their new API
					// access token
					session, _ := gSessionCache.CreateSession(user.ID)
					data.Token = session.Token
				} else { // Otherwise, if the User is linked in our database
					// Update the Facebook Link's access token in the database
					// (NOTE: We are assuming, at this point, potentially
					// dangerously, that the User definitely has a valid Facebook
					// link with the provided Facebook User ID in the database.)
					c := gDatabase.db.DB(dbDB).C("fb_links")
					query := bson.M{"fb_user_id": fbUserID}
					change := bson.M{"$set": bson.M{"fb_access_token": r.FormValue("fb_access_token")}}
					_ = c.Update(query, change)

					// Create the new Session for the user and return their new API
					// access token
					session, _ := gSessionCache.CreateSession(m["user_id"].(int))
					data.Token = session.Token
				}
			}
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		// Retrieve the User
		// (NOTE: We are, possibly dangreously, assuming that if we have a
		// valid session, a valid user definitely exists.)
		user, _, _ := gUserCache.GetUser(userID)

		// Parse out some settings from the User object
		data.ShareLocation = user.ShareLocation

		data.FriendMen = false
		data.FriendWomen = false
		data.DateMen = false
		data.DateWomen = false
		for _, element := range user.Tags {
			if element == "friends_men" {
				data.FriendMen = true
			} else if element == "friends_women" {
				data.FriendWomen = true
			} else if element == "dates_men" {
				data.DateMen = true
			} else if element == "dates_women" {
				data.DateWomen = true
			}
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		// Get the UserCache index for the User object's local representation
		_, userCacheIndex, _ := gUserCache.GetUser(userID)

		// Update the local representation of the User
		if r.FormValue("sharelocation") == "true" {
			gUserCache.Users[userCacheIndex].ShareLocation = true
		} else if r.FormValue("sharelocation") == "false" {
			gUserCache.Users[userCacheIndex].ShareLocation = false
		}

		if r.FormValue("friendmen") == "true" {
			// See if the associated tag is in the User's tags list
			found := false
			for _, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "friends_men" {
					// Stop looking
					found = true
					break
				}
			}

			// If not found, add it to the User's tags list
			if !found {
				gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags, "friends_men")
			}
		} else if r.FormValue("friendmen") == "false" {
			// See if the associated tag is in the User's tags list
			for index, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "friends_men" {
					// Remove the tag from the User's tags list
					gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags[:index], gUserCache.Users[userCacheIndex].Tags[(index+1):]...)

					// Stop looking
					break
				}
			}
		}

		if r.FormValue("friendwomen") == "true" {
			// See if the associated tag is in the User's tags list
			found := false
			for _, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "friends_women" {
					// Stop looking
					found = true
					break
				}
			}

			// If not found, add it to the User's tags list
			if !found {
				gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags, "friends_women")
			}
		} else if r.FormValue("friendwomen") == "false" {
			// See if the associated tag is in the User's tags list
			for index, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "friends_women" {
					// Remove the tag from the User's tags list
					gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags[:index], gUserCache.Users[userCacheIndex].Tags[(index+1):]...)

					// Stop looking
					break
				}
			}
		}

		if r.FormValue("datemen") == "true" {
			// See if the associated tag is in the User's tags list
			found := false
			for _, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "dates_men" {
					// Stop looking
					found = true
					break
				}
			}

			// If not found, add it to the User's tags list
			if !found {
				gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags, "dates_men")
			}
		} else if r.FormValue("datemen") == "false" {
			// See if the associated tag is in the User's tags list
			for index, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "dates_men" {
					// Remove the tag from the User's tags list
					gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags[:index], gUserCache.Users[userCacheIndex].Tags[(index+1):]...)

					// Stop looking
					break
				}
			}
		}

		if r.FormValue("datewomen") == "true" {
			// See if the associated tag is in the User's tags list
			found := false
			for _, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "dates_women" {
					// Stop looking
					found = true
					break
				}
			}

			// If not found, add it to the User's tags list
			if !found {
				gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags, "dates_women")
			}
		} else if r.FormValue("datewomen") == "false" {
			// See if the associated tag is in the User's tags list
			for index, element := range gUserCache.Users[userCacheIndex].Tags {
				if element == "dates_women" {
					// Remove the tag from the User's tags list
					gUserCache.Users[userCacheIndex].Tags = append(gUserCache.Users[userCacheIndex].Tags[:index], gUserCache.Users[userCacheIndex].Tags[(index+1):]...)

					// Stop looking
					break
				}
			}
		}

		// Push the local representation of the User to the database
		gUserCache.Users[userCacheIndex].Push()
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
	var success = Success{Success: true, Error: ""}
	var data User
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		data, _, _ = gUserCache.GetUser(userID)
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars

		// Get the User's position in the local cache
		_, userCacheIndex, _ := gUserCache.GetUser(userID)

		// Parse the recieved values into the current app User's local object
		r.ParseForm()
		for key, values := range r.Form {
			for _, value := range values {
				if key == "name" {
					gUserCache.Users[userCacheIndex].Name = value
				} else if key == "age" {
					if num, err := strconv.Atoi(value); err == nil {
						gUserCache.Users[userCacheIndex].Age = num
					}
				} else if key == "interests" {
					gUserCache.Users[userCacheIndex].Interests = map[string]int{}
					json.Unmarshal([]byte(value), &gUserCache.Users[userCacheIndex].Interests)
				} else if key == "tags" {
					gUserCache.Users[userCacheIndex].Tags = []string{}
					_ = json.Unmarshal([]byte(value), &gUserCache.Users[userCacheIndex].Tags)
				} else if key == "bio" {
					gUserCache.Users[userCacheIndex].Bio = value
				} else if key == "images" {
					gUserCache.Users[userCacheIndex].Images = []string{}
					_ = json.Unmarshal([]byte(value), &gUserCache.Users[userCacheIndex].Images)
				} else if key == "latitude" {
					if num, err := strconv.ParseFloat(value, 32); err == nil {
						gUserCache.Users[userCacheIndex].Latitude = float32(num)
					}
				} else if key == "longitude" {
					if num, err := strconv.ParseFloat(value, 32); err == nil {
						gUserCache.Users[userCacheIndex].Longitude = float32(num)
					}
				} else if key == "last_active" {
					gUserCache.Users[userCacheIndex].LastActive = value
				}
			}
		}

		// Push the updated local object into the database
		gUserCache.Users[userCacheIndex].Push()
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars

		// Delete the User from the local cache and the database (WARNING: This
		// is as final as it gets. The acount will be gone after this!)
		gUserCache.DeleteUser(userID)
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		// Update the local cache of the User's matches
		_, userCacheIndex, _ := gUserCache.GetUser(userID)
		gUserCache.Users[userCacheIndex].PullMatches()

		// Retrieve the app User's Matches
		data.Matches = gUserCache.Users[userCacheIndex].Matches
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if matchID, err := strconv.Atoi(vars["match_id"]); err == nil {
			// Update the local cache of the User's matches
			_, userCacheIndex, _ := gUserCache.GetUser(userID)
			gUserCache.Users[userCacheIndex].PullMatches()

			// Retrieve the requested Match from the User
			if match, err := gUserCache.Users[userCacheIndex].GetMatch(matchID); err == nil {
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else if r.FormValue("message") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'message' paramater must be provided in POST data."
	} else {
		if matchID, err := strconv.Atoi(vars["match_id"]); err == nil {
			user, userCacheIndex, _ := gUserCache.GetUser(userID)
			if index, err := user.GetMatchIndex(matchID); err == nil {
				// Create the new message
				message := Message{
					ID:           len(gUserCache.Users[userCacheIndex].Matches[index].Messages),
					AuthorID:     userID,
					Message:      r.FormValue("message"),
					Date:         time.Now().String(),
					Participants: gUserCache.Users[userCacheIndex].Matches[index].Participants,
				}

				// Append it to the list of Messages
				gUserCache.Users[userCacheIndex].Matches[index].PutMessage(message)
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if matchID, err := strconv.Atoi(vars["match_id"]); err == nil {
			// Retrieve the User and update their Matches
			_, userCacheIndex, _ := gUserCache.GetUser(userID)
			gUserCache.Users[userCacheIndex].PullMatches()

			// Retrieve the Messages from the specified Match
			if matchIndex, err := gUserCache.Users[userCacheIndex].GetMatchIndex(matchID); err == nil {
				// Update the Match's Messages
				gUserCache.Users[userCacheIndex].Matches[matchIndex].PullMessages()

				// Return the Match's Messages
				data.Messages = gUserCache.Users[userCacheIndex].Matches[matchIndex].Messages
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if matchID, err := strconv.Atoi(vars["match_id"]); err == nil {
			// Retrieve the User and update their Matches
			_, userCacheIndex, _ := gUserCache.GetUser(userID)
			gUserCache.Users[userCacheIndex].PullMatches()

			// Retrieve the specified Message of the specified Match
			if matchIndex, err := gUserCache.Users[userCacheIndex].GetMatchIndex(matchID); err == nil {
				if messageID, err := strconv.Atoi(vars["message_id"]); err == nil {
					// Update the Match's Messages
					gUserCache.Users[userCacheIndex].Matches[matchIndex].PullMessages()

					// Get the specified Message
					if message, err := gUserCache.Users[userCacheIndex].Matches[matchIndex].GetMessage(messageID); err == nil {
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if matchID, err := strconv.Atoi(vars["match_id"]); err == nil {
			// Retrieve the User and update their Matches
			_, userCacheIndex, _ := gUserCache.GetUser(userID)
			gUserCache.Users[userCacheIndex].PullMatches()

			// Retrieve the specified Message of the specified Match
			if matchIndex, err := gUserCache.Users[userCacheIndex].GetMatchIndex(matchID); err == nil {
				if messageID, err := strconv.Atoi(vars["message_id"]); err == nil {
					// Update the Match's Messages
					gUserCache.Users[userCacheIndex].Matches[matchIndex].PullMessages()

					// Retrieve the specified Messages
					if index, err := gUserCache.Users[userCacheIndex].Matches[matchIndex].GetMessageIndex(messageID); err == nil {
						data.Messages = gUserCache.Users[userCacheIndex].Matches[matchIndex].Messages[(index + 1):]
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		_, userCacheIndex, _ := gUserCache.GetUser(userID)

		// Create a new MD5 hasher
		hasher := md5.New()

		// Retrieve the body content from the HTTP request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			success.Success = false
			success.Error = "Failed to proccess HTTP body."
		}

		// Calculate hash of body content
		hash := hasher.Sum(body)

		// Create new File struct so we can put it in the database
		entry := File{
			ID:     bson.NewObjectId(),
			Data:   body,
			Type:   "image/jpeg",
			Length: len(body),
			MD5:    hash,
		}

		// Push the entry into the database
		c := gDatabase.db.DB(dbDB).C("files")
		c.Insert(entry)

		// Add the new URL to the User's object and push it to the database
		// (TODO: You should really delete the previous image from the database
		// it is getting overwritten.)
		imageIndex, _ := strconv.Atoi(vars["image_id"])
		imageURL := (gAPIURL + "/file/" + entry.ID.Hex())

		if (len(gUserCache.Users[userCacheIndex].Images) - 1) < imageIndex {
			gUserCache.Users[userCacheIndex].Images = append(gUserCache.Users[userCacheIndex].Images, imageURL)
		} else {
			gUserCache.Users[userCacheIndex].Images[imageIndex] = imageURL
		}

		gUserCache.Users[userCacheIndex].Push()
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if _, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		if id, err := strconv.Atoi(vars["user_id"]); err == nil {
			// Attempt to get User from the database
			if data.User, _, err = gUserCache.GetUser(id); err != nil {
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
	var success = Success{Success: true, Error: ""}
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else if r.FormValue("feeling") != "like" && r.FormValue("feeling") != "dislike" {
		success.Success = false
		success.Error = "Invalid API call. 'feeling' paramater must either be 'like' or 'dislike'."
	} else {
		if otherUserID, err := strconv.Atoi(vars["user_id"]); err == nil {
			user, _, _ := gUserCache.GetUser(userID)
			if otherUser, _, err := gUserCache.GetUser(otherUserID); err == nil {
				if r.FormValue("feeling") == "like" {
					// Switch to the "likes" database
					c := gDatabase.db.DB(dbDB).C("likes")

					// Get the latest Like from the database so that we know
					// what the ID should be for this Like
					var count int
					var err error

					if count, err = c.Count(); err != nil {
						success.Success = false
						success.Error = "Failed to add like."
					}

					// Create the new Like
					like := Like{
						ID:      count,
						LikerID: user.ID,
						LikeeID: otherUser.ID,
					}

					// Push the new Like up to the database
					if err = c.Insert(&like); err != nil {
						success.Success = false
						success.Error = "Failed to add like."
					}

					// (TODO: Add this like to any local caches.)
				} else if r.FormValue("feeling") != "dislike" {
					// Remove any likes for the specified User by the User
					c := gDatabase.db.DB(dbDB).C("likes")
					if err := c.Remove(bson.M{"liker_id": userID, "likee_id": otherUserID}); err != nil {
						success.Success = false
						success.Error = "Failed to remove any specified likes."
					}

					// (TODO: Remove this like from any local caches.)
				}
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
	var success = Success{Success: true, Error: ""}
	var data GenericData
	var returnData ReturnData

	// Process the API call
	if r.URL.Query().Get("token") == "" {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater is required."
	} else if userID, err := gSessionCache.CheckSession(r.URL.Query().Get("token")); err != nil {
		success.Success = false
		success.Error = "Invalid API call. 'token' paramater must be a valid token."
	} else {
		var _ = vars
		var users []User
		user, userCacheIndex, _ := gUserCache.GetUser(userID)

		// Switch to the "users" collection
		c := gDatabase.db.DB(dbDB).C("users")

		// Match users based on location, age, interests, and skill levels
		// (TODO: This algorithm should be worked on and enhanced. It is not
		// "smart" right now. In the future, if there are no users in the
		// immediate location, it should look for users slightly farther away
		// (as an example). We also need to filter on tags (e.g. dating men,
		// dating women, etc.). We also need distance and age to be
		// configurable.)
		{
			// Initialize the output struct
			data.PotentialUserIDs = []int{}

			// Calculate the maximumim and minimum longitude and latitude of
			// the potential users
			var maxLatitude, minLatitude, maxLongitude, minLongitude float32
			distance := (float32)(15.0 / 69.0)

			maxLatitude = (user.Latitude + distance)
			minLatitude = (user.Latitude - distance)
			maxLongitude = (user.Longitude + distance)
			minLongitude = (user.Longitude - distance)

			// Calculate the maximum and minimum age
			years := 10

			maxAge := (user.Age + years)
			minAge := (user.Age - years)

			// Build up the query and find the potential Users
			query := bson.M{}

			query["latitude"] = bson.M{"$lte": maxLatitude, "$gte": minLatitude}
			query["longitude"] = bson.M{"$lte": maxLongitude, "$gte": minLongitude}

			query["age"] = bson.M{"$lte": maxAge, "$gte": minAge}

			if len(user.Interests) > 0 {
				queryInterests := []bson.M{}
				for key := range user.Interests {
					queryInterests = append(queryInterests, bson.M{"interests." + key: bson.M{"$exists": true}})
				}
				query["$or"] = queryInterests
			}

			if err := c.Find(query).All(&users); err == nil {
				// Update the User's Matches so that we can check against them
				// so that we don't return potential Users that have already
				// been successfully matched
				gUserCache.Users[userCacheIndex].PullMatches()

				// Pack the potential User IDs into the output struct
				for _, element := range users {
					// Filter out any Users that are already matched or liked, and filter out self
					if element.ID != userID && !gUserCache.Users[userCacheIndex].IsMatchedWith(element.ID) && !gUserCache.Users[userCacheIndex].CurrentlyLikes(element.ID) {
						data.PotentialUserIDs = append(data.PotentialUserIDs, element.ID)
					}
				}
			} else {
				success.Success = false
				success.Error = "Failed to find any users."
			}
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

// EndpointGETFileID handles the "GET /file/{file_id}" API endpoint.
func EndpointGETFileID(w http.ResponseWriter, r *http.Request) {
	// Retrieve the variables from the endpoint
	vars := mux.Vars(r)

	// Process the API call
	{
		// Switch to the "files" collection
		c := gDatabase.db.DB(dbDB).C("files")

		// Find the file
		var file File
		if err := c.Find(bson.M{"_id": bson.ObjectIdHex(vars["file_id"])}).One(&file); err == nil {
			// Write the HTTP header for the response
			w.Header().Set("Content-Type", file.Type)
			w.Header().Set("Content-Length", strconv.Itoa(file.Length))
			w.WriteHeader(http.StatusOK)

			// Write the file's data
			w.Write(file.Data)
		} else {
			log.Printf("There was an issue finding the file with the requested ID (%s) in the database.", vars["file_id"])
		}
	}
}
