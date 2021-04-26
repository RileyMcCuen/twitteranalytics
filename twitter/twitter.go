package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"strconv"

	"cloud.google.com/go/pubsub"

	"cloud.google.com/go/storage"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type (
	// TwitterCredentials data structore for twitter api credentials
	TwitterCredentials struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}

	// CleanDocument contains a list of tweets betweeen EarliestTweetID
	// LastTweetID for user with UserID that is derived from Username.
	CleanDocument struct {
		Username                             string
		UserID, LastTweetID, EarliestTweetID int64
		Tweets                               []string
	}

	// FetchMessage contains the data necessary to run a fetch using the Twitter
	// API to get tweets.
	FetchMessage struct {
		Username string
		UserID   int64
	}

	MessageHandler func(string, int64) error
)

// MaxResults is the maximum number of results that are retrieved in each request
const MaxResults = 200

// PTrue returns a pointer to a bool with value of true
func PTrue() *bool {
	ret := true
	return &ret
}

// PTrue returns a pointer to a bool with value of false
func PFalse() *bool {
	ret := false
	return &ret
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

// getTweets performs calls the Twitter API to get tweets fitting parameters
// specified in data. The tweets are passed into the tweets channels.
func getTweets(client *twitter.Client, username string, userID int64) (*CleanDocument, error) {
	var resp []twitter.Tweet
	var err error
	count, doc := 0, &CleanDocument{
		Username:        username,
		UserID:          userID,
		LastTweetID:     0,
		EarliestTweetID: math.MaxInt64,
		Tweets:          make([]string, 0),
	}
	for ok := true; ok; ok = len(resp) > 0 {
		resp, _, err = client.Timelines.UserTimeline(&twitter.UserTimelineParams{
			UserID:          userID,
			Count:           MaxResults,
			TrimUser:        PTrue(),
			ExcludeReplies:  PTrue(),
			IncludeRetweets: PFalse(),
		})
		count += len(resp)
		// Create better error reporting mechanism
		if err != nil {
			log.Printf("Encountered error when getting tweets for %s: %v", username, err)
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

// unmarshal gets the query parameters from the get request and returns them if
// they are all valid, otherwise it returns an error
func unmarshal(values url.Values) (string, int64, error) {
	username := values.Get("name")
	if username == "" {
		return "", 0, errors.New("username was empty, but should not have been")
	}
	rawUserID := values.Get("id")
	if rawUserID == "" {
		return "", 0, errors.New("id was empty, but should not have been")
	}
	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("%v; userID was not a valid int64: (%s)", err, rawUserID)
	}
	return username, userID, nil
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

// Subscribe pulls requests from a subscription and handles them whenever they
// are recieved.
func Subscribe(sub *pubsub.Subscription, messageHandler MessageHandler, quit chan bool) error {
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
		if err := messageHandler("", id); err != nil {
			log.Println(err)
			m.Nack()
		} else {
			fmt.Println("recieved a sub message")
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
		log.Println(err)
		cancel()
		return err
	}
	// If quit is called then stop getting messages from the subscription
	go func() {
		log.Println("quit")
		<-quit
		cancel()
	}()
	return nil
}

// MessageHandlerHO returns a message handler that can handle a username and
// userid, fetch messages, store the messages in the cloud, then publish
// another pubsub message.
//TODO: I removed the user id parameter because we get user id in webserver container. I made it pass the id instead so we never handle a username here
func MessageHandlerHO(tClient *twitter.Client, topic *pubsub.Topic, bucket *storage.BucketHandle) MessageHandler {
	return func(name string, id int64) error {
		// Get the list of tweets
		tweets, err := getTweets(tClient, name, id)
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
		fmt.Println("published to documents")
		return nil
	}
}
