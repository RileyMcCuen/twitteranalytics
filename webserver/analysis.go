package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type (
	// DocumentMetaData is all of the data aside form tweets describing an entity
	// in the datastore.
	DocumentMetaData struct {
		UserID, LastTweetID, EarliestTweetID int64
	}

	// AnalysedDocument is the actual entity that is stored in the datastore, all of
	// the tweets have been converted to scores. There is no way to get the original
	// tweet back from the score at this point.
	AnalysedDocument struct {
		DocumentMetaData
		TweetScores                    []int
		PositiveTweets, NegativeTweets int
		AverageScore                   float64
	}

	// TwitterCredentials data structore for twitter api credentials
	TwitterCredentials struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}

	// FetchMessage contains the data necessary to run a fetch using the Twitter
	// API to get tweets.
	FetchMessage struct {
		Username string
		UserID   int64
	}
)

// GetTwitterClient function to authorize twitter api and create a client
func GetTwitterClient(credentials *TwitterCredentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(credentials.ConsumerKey, credentials.ConsumerSecret)
	token := oauth1.NewToken(credentials.AccessToken, credentials.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}
	// _ = user //silencing an error for now
	return client, nil
}

// writeError returns a plaintext message to the caller including the location
// where the error occurred, what the error was, and a BadRequest status.
func writeError(location string, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("Location: %s, Error: %v", location, err.Error())))
}

// unmarshal gets the query parameters from the get request and returns them if
// they are all valid, otherwise it returns an error
func unmarshal(values url.Values) (string, error) {
	username := values.Get("name")
	if username == "" {
		return "", errors.New("username was empty, but should not have been")
	}
	return username, nil
}

// getUser get user info based on the user input (screen name)
func getUser(client *twitter.Client, username string) (*twitter.User, error) {
	includeEntities := true
	searchParams := twitter.UserSearchParams{
		Query:           username,
		Page:            1,
		Count:           1,
		IncludeEntities: &includeEntities,
	}
	users, _, err := client.Users.Search(username, &searchParams)
	if err != nil {
		return nil, err
	}
	if len(users) < 1 {
		return nil, errors.New("no users were found with that username")
	}
	return &users[0], nil
}

func getData(username string, userID int64, ds *datastore.Client, topic *pubsub.Topic) (interface{}, error) {
	doc := &AnalysedDocument{}
	if err := ds.Get(context.Background(), datastore.IDKey("User", userID, nil), doc); err != nil {
		fm := &FetchMessage{
			Username: username,
			UserID:   userID,
		}
		message, err := json.Marshal(fm)
		if err != nil {
			return nil, err
		}
		res := topic.Publish(context.Background(), &pubsub.Message{
			Data: message,
		})

		if _, err := res.Get(context.Background()); err != nil {
			return nil, err
		}
		return struct{ Message string }{Message: "This user has not been analysed yet, they have been submitted to be analysed. Check back later to see more about them."}, nil
	}
	return *doc, nil
}

// GetAnalysisHO...
func GetAnalysisHO(tClient *twitter.Client, ds *datastore.Client, topic *pubsub.Topic) http.HandlerFunc {
	// GetAnalysis either gets the analysis data for a twitter user who has
	// already been processed, or it sends a request to start analysing them.
	return func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			// Must be a GET request
			writeError("Method", errors.New("/api/analysis only accepts GET requests"), w)
			return
		}
		// Unmarshal the request into name variable
		name, err := unmarshal(r.URL.Query())
		if err != nil {
			writeError("Unmarshal", err, w)
		}
		user, err := getUser(tClient, name)
		if err != nil {
			writeError("User", err, w)
		}
		data, err := getData(user.Name, user.ID, ds, topic)
		if err != nil {
			writeError("Data", err, w)
		}
		// Marshal the analysis data into JSON format for transport
		ret, err := json.Marshal(data)
		if err != nil {
			writeError("Marshal", err, w)
			return
		}
		// Send the json data to the requester
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(ret)
	}
}
