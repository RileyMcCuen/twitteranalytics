package main

import (
	"fmt"
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

//getTwitterClient function to authorize twitter api and create a client
func getTwitterClient(credentials *TwitterCredentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(credentials.ConsumerKey, credentials.ConsumerSecret)
	token := oauth1.NewToken(credentials.AccessToken, credentials.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}
	//log.Printf("user's account:\n%+v\n", user)
	_ = user //silencing an error for now
	return client, nil
}

//getUser get user info based on the user input (screen name)
func getUser(client twitter.Client, username string) (*twitter.User, error) {
	includeEntities := true
	searchParams := twitter.UserSearchParams{
		Query:           username,
		Page:            1,
		Count:           1,
		IncludeEntities: &includeEntities,
	}
	users, _, err := client.Users.Search(username, &searchParams)
	if err != nil {
		fmt.Println("error finding user")
		return nil, err
	}
	fmt.Println(users)
	return &users[0], nil
}

// Tweets gets a list of tweets specified by data
func Tweets(data *GetAnalysisData) ([]string, error) {
	tweets := make([]string, 0)
	// TODO: get a list of tweets from twitter from the last data.Days and
	//       tweeted by data.Username, put them in tweets array.
	return tweets, nil
}
