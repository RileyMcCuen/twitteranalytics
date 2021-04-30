package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
)

var envVarNames = []string{
	"GOOGLE_APPLICATION_CREDENTIALS",
	"BUCKET",
	"PROJECT_ID",
}

const (
	evGoogleApplicationCredentials = iota
	evBucket
	evProjectID
)

type (
	// Changes is the flag in the database indicating whether new data has been
	// added since the last time Changes was checked and toggled.
	Changes struct{ ChangesMade bool }

	// User is a username and user id of a user that has been analysed.
	User struct {
		Username string
		UserID   int64
	}
)

func (u User) String() string {
	return fmt.Sprintf("%-d: %s", u.UserID, u.Username)
}

const (
	changesKind   = "Changes"
	userKind      = "User"
	usernameField = "Username"
	userIDField   = "UserID"
	objectKey     = "name-index.json"

	changesKeyID = 1
	limit        = 100000
)

// changesKey is the key to an entity that indicates whether any updates to the database have happened
var changesKey = datastore.IDKey(changesKind, changesKeyID, nil)

// InitDatastore intializes the database client
func InitDatastore() *datastore.Client {
	store, err := datastore.NewClient(context.Background(), os.Getenv(envVarNames[evProjectID]))
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
	bucket := client.Bucket(os.Getenv(envVarNames[evBucket]))
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
	for _, envVar := range envVarNames {
		if _, ok := os.LookupEnv(envVar); !ok {
			log.Fatalf("Missing environment variable: %s\n", envVar)
		}
	}
}

// InitLibs prepares external libraries with credentials to make API calls.
func InitLibs() (*datastore.Client, *storage.BucketHandle) {
	VerifyEnvironment()
	return InitDatastore(), InitStorage()
}

// shouldUpdate returns true if changes have occurred in the database since
// the last time it was checked.
func shouldUpdate(ds *datastore.Client) bool {
	tx, err := ds.NewTransaction(context.Background())
	if err != nil {
		log.Println(err)
		return false
	}
	dst := &Changes{}
	if err := tx.Get(changesKey, dst); err != nil {
		log.Println(err)
		tx.Rollback()
		return false
	}
	if dst.ChangesMade {
		dst.ChangesMade = false
		if _, err := tx.Put(changesKey, dst); err != nil {
			log.Println(err)
		}
		if _, err := tx.Commit(); err != nil {
			log.Println(err)
		}
		return true
	}
	return false
}

// users gets the first 100,000 (username, id) pairs from the database.
func users(ds *datastore.Client) ([]User, error) {
	dst := make([]User, 0)
	_, err := ds.GetAll(context.Background(), datastore.NewQuery(userKind).Limit(limit).Project(usernameField, userIDField).Order(usernameField), &dst)
	return dst, err
}

// store, stores the provided users in a json array in the given bucket.
func store(bucket *storage.BucketHandle, users []User) error {
	w := bucket.Object(objectKey).NewWriter(context.Background())
	return json.NewEncoder(w).Encode(users)
}

// printUsers is a debug helper method that prints out an array of users.
func printUsers(users []User) {
	for _, user := range users {
		log.Println(user)
	}
}

// main starts up the webserver.
func main() {
	// Get clients
	ds, bucket := InitLibs()
	_, err := ds.Put(context.Background(), changesKey, &Changes{ChangesMade: true})
	if err != nil {
		log.Println(err)
		return
	}
	if shouldUpdate(ds) {
		users, err := users(ds)
		if err != nil {
			log.Println(err)
		}
		printUsers(users)
		if err := store(bucket, users); err != nil {
			log.Println(err)
		}
	} else {
		log.Println("No changes detected...")
	}
}
