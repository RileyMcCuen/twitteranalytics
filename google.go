package main

import (
	"context"
	"errors"
	"log"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// analyseSentiment get sentiment scores from Google's Natural Language API for
// all of the strings input on in, and outputs the scores on out.
func analyseSentiment(client *language.Client, ctx context.Context, in chan string, out chan float64) {
	for tweet := range in {
		sentiment, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
			Document: &languagepb.Document{
				Source: &languagepb.Document_Content{
					Content: tweet,
				},
				Type: languagepb.Document_PLAIN_TEXT,
			},
			EncodingType: languagepb.EncodingType_UTF8,
		})
		if err != nil {
			log.Println(err)
		} else {
			out <- float64(sentiment.DocumentSentiment.Score)
		}
	}
	close(out)
}

// analyseCategory gets categories from Google's Natural Language API for
// all of the strings input on in, and outputs the categories on out.
func analyseCategory(client *language.Client, ctx context.Context, in chan string, out chan string) {
	for tweet := range in {
		class, err := client.ClassifyText(ctx, &languagepb.ClassifyTextRequest{
			Document: &languagepb.Document{
				Source: &languagepb.Document_Content{
					Content: tweet,
				},
				Type: languagepb.Document_PLAIN_TEXT,
			},
		})
		if err != nil {
			log.Println(err)
		} else {
			for _, cat := range class.Categories {
				out <- cat.Name
			}
		}
	}
	close(out)
}

// collectData gets all of the categories and scores from the output channels
// and fills in all of the fields of an AnalysedData object. When both output
// channels are closes a quit message is sent on the quit channel.
func collectData(data *AnalysedData, sentOut chan float64, catOut chan string, quit chan error) {
	for sentOut != nil && catOut != nil {
		select {
		case sent, ok := <-sentOut:
			if ok {
				data.MeanScore += sent
				data.Scores = append(data.Scores, ScoreCount{Score: sent, Count: 1})
			} else {
				sentOut = nil
			}
		case cat, ok := <-catOut:
			if ok {
				data.Topics = append(data.Topics, TopicCount{Topic: cat, Count: 1})
			} else {
				catOut = nil
			}
		}
	}

	if len(data.Scores) > 0 {
		data.
			sortSentiments(). // sort to calculate median
			count().
			calculateMedian().
			calculateMean().
			collectScores().
			sortSentiments(). // sort so returned values are in order after collect
			calculateScoreDist().
			collectTopics().
			sortTopics()

		quit <- nil
	} else {
		quit <- errors.New("there were no tweets in the last week for that person")
	}

	close(quit)
}

// Analyse uses the Google Natural Language API to analyse a stream of tweets
// passed in on the tweets channel. The results and an error if any where found
// are returned.
func Analyse(client *language.Client, tweets chan string) (*AnalysedData, error) {
	// Make the input channels for the goroutines
	sentIn, catIn := make(chan string, MaxResults), make(chan string, MaxResults)
	// Make the ouput channels for the goroutines
	sentOut, catOut := make(chan float64, MaxResults), make(chan string, MaxResults)
	// Make other variables for requests
	ctx, data, quit := context.Background(), &AnalysedData{}, make(chan error, 0)

	// Start up all of the goroutines
	go analyseSentiment(client, ctx, sentIn, sentOut)
	go analyseCategory(client, ctx, catIn, catOut)
	go collectData(data, sentOut, catOut, quit)

	// Get all of the tweets and pass them in to be analysed
	for tweet := range tweets {
		sentIn <- tweet
		catIn <- tweet
	}

	// All tweets have been retrieved, close the input channels
	close(sentIn)
	close(catIn)

	return data, <-quit
}
