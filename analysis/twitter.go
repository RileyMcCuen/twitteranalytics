package main

import (
	"errors"
	"log"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// TwitterCredentials data structore for twitter api credentials
type TwitterCredentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

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
	// log.Printf("user's account:\n%+v\n", user)
	// _ = user //silencing an error for now
	return client, nil
}

// getUser get user info based on the user input (screen name)
func getUser(client *twitter.Client, username string) (*twitter.User, error) {
	includeEntities := true
	log.Println(username)
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

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const (
	MinDays    = 1
	MaxDays    = 7
	MaxResults = 200
)

func PTrue() *bool {
	ret := true
	return &ret
}

func PFalse() *bool {
	ret := false
	return &ret
}

// calculateStartTime calculates the UTC time days before right now.
func calculateStartTime(days int) time.Time {
	// TODO: actually calculate start time using days input
	return time.Now().UTC()
}

// jake and logan paul, alex jones, joe rogan,

// getTweets performs calls the Twitter API to get tweets fitting parameters
// specified in data. The tweets are passed into the tweets channels.
func getTweets(client *twitter.Client, data *GetAnalysisData, tweets chan string) {
	var resp []twitter.Tweet
	count := 0
	for ok := true; ok; ok = len(resp) > 0 {
		// VERIFY that resp is not being shadowed
		resp, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
			UserID:          data.UserID,
			Count:           MaxResults,
			TrimUser:        PTrue(),
			ExcludeReplies:  PTrue(),
			IncludeRetweets: PFalse(),
		})
		count += len(resp)
		// Create better error reporting mechanism
		if err != nil || count >= 600 {
			break
		}
		for _, tweet := range resp {
			tweets <- tweet.Text
		}
	}
	// TODO: Implement pagination to get all tweets for specified days
	// search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
	// 	Query: fmt.Sprintf("from:%s", data.Username),
	// 	Count: MaxResults,
	// })
	close(tweets)
}

// Tweets gets a list of tweets specified by data
func Tweets(client *twitter.Client, data *GetAnalysisData) (chan string, error) {
	user, err := getUser(client, data.Username)
	if err != nil {
		return nil, err
	}
	data.UserID = user.ID
	data.Days = intMin(intMax(MinDays, data.Days), MaxDays)
	tweets := make(chan string, MaxResults)
	go getTweets(client, data, tweets)
	return tweets, nil
}
