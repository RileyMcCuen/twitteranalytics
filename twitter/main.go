package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"

	"cloud.google.com/go/storage"
	"github.com/dghubble/go-twitter/twitter"
)

// InitTwitter initializes the twitter api client
func InitTwitter() *twitter.Client {
	twitterCredentials := TwitterCredentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
		ConsumerKey:       os.Getenv("API_KEY"),
		ConsumerSecret:    os.Getenv("API_SECRET_KEY"),
	}
	client, err := GetTwitterClient(&twitterCredentials)
	if err != nil {
		log.Fatalf("Failed to create Twitter client: %v\n", err)
	}
	return client
}

// InitStorage creates a client, gets a bucket handle and then verifies that the
// bucket exists.
func InitStorage() *storage.BucketHandle {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v\n", err)
	}
	bucket := client.Bucket(os.Getenv("BUCKET"))
	attrs, err := bucket.Attrs(context.Background())
	if attrs == nil {
		if err != nil {
			fmt.Println(err)
		}
		log.Fatalf("Bucket has no attributes...\n")
	}
	if err != nil {
		log.Fatalf("Could not get Bucket information: %v\n", err)
	}
	return bucket
}

// TODO: Make a pubsub for the tweets to be processed
func InitPubSub() *pubsub.Client {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Could not set up pub sub client: #{err}\n")
	}
	return client
}

// VerifyEnvironment verifies that all expected environment variables exist
func VerifyEnvironment() {
	envVariables := [...]string{
		"ACCESS_TOKEN",
		"ACCESS_TOKEN_SECRET_KEY",
		"API_KEY",
		"API_SECRET_KEY",
		"GOOGLE_APPLICATION_CREDENTIALS",
		"BUCKET",
		"ADDRESS",
		"PROJECT_ID",
		"PUB_SUB_SUBSCRIPTION_ID",
		"PUB_SUB_TOPIC_ID",
	}
	for _, envVar := range envVariables {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("twitter Missing environment variable: %s\n", envVar)
		}
	}
}

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*twitter.Client, *pubsub.Client, *storage.BucketHandle) {
	VerifyEnvironment()
	return InitTwitter(), InitPubSub(), InitStorage()
}

// ConfigurePubSub gets a subscription and topic and makes sure that both exist.
func ConfigurePubSub(psClient *pubsub.Client) (*pubsub.Subscription, *pubsub.Topic) {
	subID, topicID := os.Getenv("PUB_SUB_SUBSCRIPTION_ID"), os.Getenv("PUB_SUB_TOPIC_ID")
	sub, topic := psClient.Subscription(subID), psClient.Topic(topicID)
	if ok, err := sub.Exists(context.Background()); !ok || err != nil {
		log.Fatalf("Subscription: %s does not exist. Error: %v\n", subID, err)
	}
	if ok, err := topic.Exists(context.Background()); !ok || err != nil {
		log.Fatalf("Topic: %s does not exist. Error: %v\n", topicID, err)
	}
	return sub, topic
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// Get clients
	tClient, psClient, bucket := InitLibs()
	sub, topic := ConfigurePubSub(psClient)
	quit := make(chan bool)
	// Handle calls to the analysis endpoint
	err := Subscribe(sub, MessageHandlerHO(tClient, topic, bucket), quit)
	if err != nil {
		log.Fatalf("Encountered an error when trying to susbcribe: %v\n", err)
	}
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe(os.Getenv("ADDRESS"), nil))
}
