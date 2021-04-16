package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	language "cloud.google.com/go/language/apiv1"
	"github.com/dghubble/go-twitter/twitter"
)

type (
	// GetAnalysisData is the data used to make an analysis request
	GetAnalysisData struct {
		// Username is the username of the Twitter user to analyse
		Username string
		// Days is the number of days in the past to collect tweets from
		Days int
	}

	// AnalysedData is the result of GetAnalysisData after it is proccessed
	AnalysedData struct {
		// Scores is a list of all sentiment analysis scores
		Scores []float64
		// MeanScore and MedianScore are calculated based on Scores
		MeanScore, MedianScore float64
		// Topics are all of the topics of the tweets
		Topics []string
	}
)

func writeError(location string, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("Location: %s, Error: %v", location, err.Error())))
}

func unmarshal(values url.Values) (*GetAnalysisData, error) {
	data := &GetAnalysisData{}
	username := values.Get("name")
	if username == "" {
		return nil, errors.New("username was empty, but should not have been")
	}
	data.Username = username
	data.Days = 1
	numDays := values.Get("days")
	if numDays != "" {
		if days, err := strconv.Atoi(numDays); err == nil {
			data.Days = days
		}
	}
	return data, nil
}

func GetAnalysisHO(tClient *twitter.Client, gClient *language.Client) http.HandlerFunc {
	// GetAnalysis pulls data from twitter given a username, then passes the
	// tweets off to Google Natural Language Processing API which classifies them
	// and analyses sentiment.
	return func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			// Must be a GET request
			writeError("Check Method", errors.New("/api/analysis only accepts GET requests"), w)
			return
		}
		// Unmarshal the request into data variable
		data, err := unmarshal(r.URL.Query())
		if err != nil {
			writeError("Unmarshal Data", err, w)
		}
		// Get the list of tweets
		tweets, err := Tweets(tClient, data)
		if err != nil {
			writeError("Tweets", err, w)
			return
		}
		// Analyse the tweets
		analysedData, err := Analyse(tweets)
		if err != nil {
			writeError("Analyse", err, w)
			return
		}
		// Marshal the analysis data into JSON format for transport
		ret, err := json.Marshal(analysedData)
		if err != nil {
			writeError("Marshal Data", err, w)
			return
		}
		// Send the json data to the requester
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(ret)
	}
}
