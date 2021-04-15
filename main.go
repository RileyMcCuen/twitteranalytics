package main

import (
	"log"
	"net/http"
)

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() {
	// TODO: Use environment variables to initialize twitter and google libs
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// TODO: Use environment variables for address and credentials
	InitLibs()
	// Handle requests for static files
	http.Handle("/static", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// Handle calls to the analysis endpoint
	http.HandleFunc("/api/analysis", GetAnalysis)
	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe("0.0.0.0:80", nil))
}
