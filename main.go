package main

import (
	"context"
	"log"
	"net/http"
	"os"

	language "cloud.google.com/go/language/apiv1"

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

// InitGoogle initializes the google natural language api client
func InitGoogle() *language.Client {
	ctx := context.Background()
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Google client: %v", err)
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
		"ADDRESS",
	}
	for _, envVar := range envVariables {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Missing environment variable: %s\n", envVar)
		}
	}
}

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*twitter.Client, *language.Client) {
	VerifyEnvironment()
	return InitTwitter(), InitGoogle()
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// Get clients
	tClient, gClient := InitLibs()

	// Handle requests for static files
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// Handle calls to the analysis endpoint
	http.HandleFunc("/api/analysis", GetAnalysisHO(tClient, gClient))
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe(os.Getenv("ADDRESS"), nil))
}
