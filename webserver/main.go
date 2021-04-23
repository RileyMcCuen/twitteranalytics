package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	objectKey = "name-index.json"
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

// InitDatastore intializes the database client
func InitDatastore() *datastore.Client {
	store, err := datastore.NewClient(context.Background(), os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Could not get datastore client: %v\n", err)
	}
	return store
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
		log.Fatalf("Bucket has not attributes...\n")
	}
	if err != nil {
		log.Fatalf("Could not get Bucket information: %v\n", err)
	}
	return bucket
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
		"TWITTER_SERVICE_BASE_PATH",
		"PROJECT_ID",
	}
	for _, envVar := range envVariables {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Missing environment variable: %s\n", envVar)
		}
	}
}

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*twitter.Client, *datastore.Client, *storage.BucketHandle) {
	VerifyEnvironment()
	return InitTwitter(), InitDatastore(), InitStorage()
}

// Users gets a list of users from a file in the bucket
func UsersHO(bucket *storage.BucketHandle) http.HandlerFunc {
	obj := bucket.Object(objectKey)
	return func(w http.ResponseWriter, r *http.Request) {
		dataReader, err := obj.NewReader(context.Background())
		if err != nil {
			w.Write([]byte("{\"Message\":\"Could not get index from bucket. There are likely no users in it yet. Make a new Query!\""))
		}
		io.Copy(w, dataReader)
	}
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// Get clients
	tClient, ds, bucket := InitLibs()

	// Handle requests for static files
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// Handle calls to the analysis endpoint
	http.HandleFunc("/api/analyse", GetAnalysisHO(tClient, ds, os.Getenv("TWITTER_SERVICE_BASE_PATH")))
	// Handle calls to get list of users that have already been analysed
	http.HandleFunc("/api/user", UsersHO(bucket))
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe(os.Getenv("ADDRESS"), nil))
}
