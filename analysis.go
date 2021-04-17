package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
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

	ScoreCount struct {
		Score float64
		Count int
	}

	SortableSentiments []ScoreCount

	TopicCount struct {
		Topic string
		Count int
	}

	SortableTopics []TopicCount

	SentimentDist struct {
		Negative, Neutral, Positive int
	}

	// AnalysedData is the result of GetAnalysisData after it is proccessed
	AnalysedData struct {
		// Scores is a list of all sentiment analysis scores
		Scores SortableSentiments
		// MeanScore and MedianScore are calculated based on Scores
		MeanScore, MedianScore float64
		// Count is the number of tweets that were analysed
		Count int
		// SentimentDist contains the number of elements in one of three categories
		ScoreDist SentimentDist
		// Topics are all of the topics of the tweets
		Topics SortableTopics
	}
)

func (ss SortableSentiments) Len() int {
	return len(ss)
}

func (ss SortableSentiments) Less(i, j int) bool {
	return ss[i].Score < ss[j].Score
}

func (ss SortableSentiments) Swap(i, j int) {
	tmp := ss[i]
	ss[i] = ss[j]
	ss[j] = tmp
}

func (sc SortableTopics) Len() int {
	return len(sc)
}

func (sc SortableTopics) Less(i, j int) bool {
	return sc[i].Count < sc[j].Count
}

func (sc SortableTopics) Swap(i, j int) {
	tmp := sc[i]
	sc[i] = sc[j]
	sc[j] = tmp
}

func (d *AnalysedData) sortSentiments() *AnalysedData {
	sort.Sort(d.Scores)
	return d
}

func (d *AnalysedData) sortTopics() *AnalysedData {
	sort.Sort(d.Topics)
	return d
}

func (d *AnalysedData) count() *AnalysedData {
	d.Count = d.Scores.Len()
	return d
}

func (d *AnalysedData) calculateMedian() *AnalysedData {
	mid := len(d.Scores) / 2
	if len(d.Scores)%2 == 0 {
		d.MedianScore = (d.Scores[mid-1].Score + d.Scores[mid].Score) / 2
	} else {
		d.MeanScore = d.Scores[mid].Score
	}
	return d
}

func (d *AnalysedData) calculateMean() *AnalysedData {
	d.MeanScore /= float64(d.Scores.Len())
	return d
}

func (d *AnalysedData) collectScores() *AnalysedData {
	scores := make(map[float64]int)
	for _, score := range d.Scores {
		count, ok := scores[score.Score]
		if !ok {
			count = 0
		}
		scores[score.Score] = count + score.Count
	}
	scoresArr := make([]ScoreCount, len(scores))
	i := 0
	for score, count := range scores {
		scoresArr[i] = ScoreCount{Score: score, Count: count}
		i++
	}

	d.Scores = scoresArr
	return d
}

func (d *AnalysedData) collectTopics() *AnalysedData {
	topics := make(map[string]int)
	for _, topic := range d.Topics {
		count, ok := topics[topic.Topic]
		if !ok {
			count = 0
		}
		topics[topic.Topic] = count + topic.Count
	}
	topicsArr := make([]TopicCount, len(topics))
	i := 0
	for topic, count := range topics {
		topicsArr[i] = TopicCount{Topic: topic, Count: count}
		i++
	}

	d.Topics = topicsArr
	return d
}

func (d *AnalysedData) calculateScoreDist() *AnalysedData {
	d.ScoreDist = SentimentDist{}
	for _, score := range d.Scores {
		if score.Score <= -0.5 {
			d.ScoreDist.Negative += score.Count
		} else if score.Score >= 0.5 {
			d.ScoreDist.Positive += score.Count
		} else {
			d.ScoreDist.Neutral += score.Count
		}
	}
	return d
}

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
			writeError("Method", errors.New("/api/analysis only accepts GET requests"), w)
			return
		}
		// Unmarshal the request into data variable
		data, err := unmarshal(r.URL.Query())
		if err != nil {
			writeError("Unmarshal", err, w)
		}
		// Get the list of tweets
		tweets, err := Tweets(tClient, data)
		if err != nil {
			writeError("Tweets", err, w)
			return
		}
		// Analyse the tweets
		analysedData, err := Analyse(gClient, tweets)
		if err != nil {
			writeError("Analyse", err, w)
			return
		}
		// Marshal the analysis data into JSON format for transport
		ret, err := json.Marshal(analysedData)
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
