package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"log"
	"net/http"
	"os"
)

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*twitter.Client, error) {
	// TODO: Use environment variables to initialize twitter and google libs
	//get a twitter client
	twitterCredentials := TwitterCredentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET_KEY"),
		ConsumerKey:       os.Getenv("API_KEY"),
		ConsumerSecret:    os.Getenv("API_SECRET_KEY"),
	}

	client, err := getTwitterClient(&twitterCredentials)
	if err != nil {
		fmt.Println("error getting twitter client")
		return nil, err
	}

	//fmt.Printf("%+v\n", client)
	return client, nil
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// TODO: Use environment variables for address and credentials
	client, err := InitLibs()
	if err != nil {
		return
	}

	//This is temporary, get user function will eventually be called based on user input
	user, err := getUser(*client, "tdhoward55")

	if err != nil {
		fmt.Println("error")
		return
	}
	fmt.Println("user: ")
	//_ = user
	fmt.Println(user.ScreenName)
	//end of get user testing

	// Handle requests for static files
	http.Handle("/static", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// Handle calls to the analysis endpoint
	http.HandleFunc("/api/analysis", GetAnalysis)
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe("0.0.0.0:80", nil))
}
