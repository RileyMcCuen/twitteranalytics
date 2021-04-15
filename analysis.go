package main

import (
	"encoding/json"
	"net/http"
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

func writeError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

// GetAnalysis pulls data from twitter given a username, then passes the
// tweets off to Google Natural Language Processing API which classifies them
// and analyses sentiment.
func GetAnalysis(w http.ResponseWriter, r *http.Request) {
	// Check request method
	if r.Method != http.MethodGet {
		// Must be a GET request
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("/api/analysis only accepts GET requests"))
		return
	}
	// Unmarshal the request into data variable
	data, decoder := &GetAnalysisData{}, json.NewDecoder(r.Body)
	if err := decoder.Decode(data); err != nil {
		// If request cannot be unmarshalled then return appropriate error
		writeError(err, w)
		return
	}
	// Get the list of tweets
	tweets, err := Tweets(data)
	if err != nil {
		writeError(err, w)
		return
	}
	// Analyse the tweets
	analysedData, err := Analyse(tweets)
	if err != nil {
		writeError(err, w)
		return
	}
	// Marshal the analysis data into JSON format for transport
	ret, err := json.Marshal(analysedData)
	if err != nil {
		writeError(err, w)
		return
	}
	// Send the json data to the requester
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(ret)
}
