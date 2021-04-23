package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/cdipaolo/sentiment"
)

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

func InitModel() sentiment.Models {
	model, err := sentiment.Restore()
	if err != nil {
		log.Fatalf("Could not restore model: %v\n", err)
	}
	return model
}

func InitDatastore() *datastore.Client {
	store, err := datastore.NewClient(context.Background(), os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Could not get datastore client: %v\n", err)
	}
	return store
}

func InitPubSubClient() *pubsub.Client {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Could not initialize pub sub client: #{err}\n")
	}
	return client
}

// TODO: Make a pubsub for the tweets to be processed

// VerifyEnvironment verifies that all expected environment variables exist
func VerifyEnvironment() {
	envVariables := [...]string{
		"GOOGLE_APPLICATION_CREDENTIALS",
		"BUCKET",
		"ADDRESS",
		"PROJECT_ID",
	}
	for _, envVar := range envVariables {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Missing environment variable: %s\n", envVar)
		}
	}
}

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*storage.BucketHandle, sentiment.Models, *datastore.Client, *pubsub.Client) {
	VerifyEnvironment()
	return InitStorage(), InitModel(), InitDatastore(), InitPubSubClient()
}

// Health is a probe endpoint, it always returns StatusOK and "Healthy".
func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

// main starts up the webserver.
func main() {
	// Get clients
	bucket, model, ds, pubSubClient := InitLibs()
	//TODO: sub here, call Analyse when the topic has been published to
	ctx := context.Background()
	//TODO: check if this is the write way to make sure topic exists
	topic, err := pubSubClient.CreateTopic(ctx, "TwitterAnalysis427")
	if err != nil {
		topic = pubSubClient.Topic("TwitterAnalysis427")
	}
	//TODO: include pod number in id?
	sub, err := pubSubClient.CreateSubscription(ctx, "analysis-subscription", pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		log.Fatalf("Error creating subscriber: #{err}]\n")
	}
	receiveErr := sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		//TODO: analyse
		file := string(message.Data)
		Analyse(bucket, model, ds, file)
		message.Ack()
	})
	//TODO: make sure this is appropriate error handling
	if receiveErr != nil {
		log.Fatalf("Error receiving from publisher: #{err}]\n")
	}

	//TODO: make this not a go routine
	//stopExecuting := make(chan bool, 0)
	//go Analyse(bucket, model, ds, stopExecuting)

	// Handle calls to the health endpoint
	http.HandleFunc("/api/health", Health)
	log.Fatal(http.ListenAndServe(os.Getenv("ADDRESS"), nil))
}
