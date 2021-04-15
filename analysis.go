package main

import (
	"encoding/json"
	"net/http"
)

// GetAnalysisData is the data used to make an analysis request
type GetAnalysisData struct {
	// Username is the username of the Twitter user to analyse
	Username string
	// Days is the number of days in the past to collect tweets from
	Days int
}

// GetAnalysis pulls data from twitter given a username, then passes the
// tweets off to Google Natural Language Processing API which classifies them
// and analyses sentiment.
func GetAnalysis(w http.ResponseWriter, r *http.Request) {
	// Unmarshal the request into data variable
	data, decoder := &GetAnalysisData{}, json.NewDecoder(r.Body)
	if err := decoder.Decode(data); err != nil {
		// If request cannot be unmarshalled then return appropriate error
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	// TODO: Fill out this function to collect data from twitter then analyse
	//       it using the google apis...
}
