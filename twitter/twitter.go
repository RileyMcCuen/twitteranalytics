package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"cloud.google.com/go/pubsub"

	"cloud.google.com/go/storage"
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

const MaxResults = 200

func PTrue() *bool {
	ret := true
	return &ret
}

func PFalse() *bool {
	ret := false
	return &ret
}

type CleanDocument struct {
	UserID, LastTweetID, EarliestTweetID int64
	Tweets                               []string
}

// jake and logan paul, alex jones, joe rogan,

// getTweets performs calls the Twitter API to get tweets fitting parameters
// specified in data. The tweets are passed into the tweets channels.
func getTweets(client *twitter.Client, userID int64) (*CleanDocument, error) {
	var resp []twitter.Tweet
	var err error
	count, doc := 0, &CleanDocument{UserID: userID, LastTweetID: 0, EarliestTweetID: math.MaxInt64, Tweets: make([]string, 0)}
	for ok := true; ok; ok = len(resp) > 0 {
		// VERIFY that resp is not being shadowed
		resp, _, err = client.Timelines.UserTimeline(&twitter.UserTimelineParams{
			UserID:          userID,
			Count:           MaxResults,
			TrimUser:        PTrue(),
			ExcludeReplies:  PTrue(),
			IncludeRetweets: PFalse(),
		})
		count += len(resp)
		// Create better error reporting mechanism
		if err != nil || count >= 600 {
			return doc, err
		}

		for _, tweet := range resp {
			if tweet.ID > doc.LastTweetID {
				doc.LastTweetID = tweet.ID
			}
			if tweet.ID < doc.EarliestTweetID {
				doc.EarliestTweetID = tweet.ID
			}
			doc.Tweets = append(doc.Tweets, tweet.Text)
		}
	}
	if doc.LastTweetID == 0 && doc.EarliestTweetID == math.MaxInt64 {
		return doc, errors.New("no tweets were found, so no analysis can be done")
	}
	return doc, nil
}

// Tweets gets a list of tweets specified by username
func Tweets(client *twitter.Client, userID int64) (*CleanDocument, error) {
	//user, err := getUser(client, username)
	//if err != nil {
	//	return nil, err
	//}
	return getTweets(client, userID)
}

func writeError(location string, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("Location: %s, Error: %v", location, err.Error())))
}

func unmarshal(values url.Values) (string, error) {
	username := values.Get("name")
	if username == "" {
		return username, errors.New("username was empty, but should not have been")
	}
	return username, nil
}

// storeTweets writes the document containing tweet information in a document
func storeTweets(bucket *storage.BucketHandle, doc *CleanDocument) (string, error) {
	fileName := fmt.Sprintf("%d-%d.json", doc.UserID, doc.LastTweetID)
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	w := bucket.Object(fileName).NewWriter(context.Background())
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return "", err
	}
	return fileName, nil
}

type FetchMessage struct {
	Username string
	UserID   int64
}

// Subscribe pulls requests from a subscription and handles then whenever they
// are recieved.
func Subscribe(sub *pubsub.Subscription, messageHandler func(int64) error, quit chan bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: I don't think this should be an if
	if err := sub.Receive(ctx, func(c context.Context, m *pubsub.Message) {
		idString := string(m.Data)
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			log.Println(err)
			m.Nack()
		}
		log.Println(id)
		//TODO: call message handler
		if err := messageHandler(id); err != nil {
			log.Println(err)
			m.Nack()
		} else {
			m.Ack()
		}
		//fm := FetchMessage{}
		////TODO: I think this will throw an error, replaced with the above code
		//if err := json.Unmarshal(m.Data, &fm); err != nil {
		//	log.Println(err)
		//	m.Nack()
		//}
		//if err := messageHandler(fm.Username, fm.UserID); err != nil {
		//	log.Println(err)
		//	m.Nack()
		//} else {
		//	m.Ack()
		//}
	}); err != nil {
		cancel()
		return err
	}
	// If quit is called then stop getting messages from the subscription
	go func() {
		<-quit
		cancel()
	}()
	return nil
}

// MessageHandlerHO returns a message handler that can handle a username and
// userid, fetch messages, store the messages in the cloud, then publish
// another pubsub message.
//TODO: I removed the user id parameter because we get user id in webserver container. I made it pass the id instead so we never handle a username here
func MessageHandlerHO(tClient *twitter.Client, topic *pubsub.Topic, bucket *storage.BucketHandle) func(int64) error {
	return func(userID int64) error {
		// Get the list of tweets
		tweets, err := Tweets(tClient, userID)
		if err != nil {
			return err
		}
		message, err := storeTweets(bucket, tweets)
		if err != nil {
			return err
		}

		res := topic.Publish(context.Background(), &pubsub.Message{
			Data: []byte(message),
		})

		if _, err := res.Get(context.Background()); err != nil {
			return err
		}

		return nil
	}
}
