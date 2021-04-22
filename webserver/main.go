package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
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

// InitDatastore intializes the database client
func InitDatastore() *datastore.Client {
	store, err := datastore.NewClient(context.Background(), os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Could not get datastore client: %v\n", err)
	}
	return store
}

// VerifyEnvironment verifies that all expected environment variables exist
func VerifyEnvironment() {
	envVariables := [...]string{
		"ACCESS_TOKEN",
		"ACCESS_TOKEN_SECRET_KEY",
		"API_KEY",
		"API_SECRET_KEY",
		"GOOGLE_APPLICATION_CREDENTIALS",
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
func InitLibs() (*twitter.Client, *datastore.Client) {
	VerifyEnvironment()
	return InitTwitter(), InitDatastore()
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// Get clients
	tClient, ds := InitLibs()

	// Handle requests for static files
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// Handle calls to the analysis endpoint
	http.HandleFunc("/api/analyse", GetAnalysisHO(tClient, ds, os.Getenv("TWITTER_SERVICE_BASE_PATH")))
	// Handle calls to get users that have already been analysed
	// http.HandleFunc("/api/analysed", TODO)
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe(os.Getenv("ADDRESS"), nil))
}
