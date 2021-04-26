package main

import (
	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"github.com/cdipaolo/sentiment"
	"log"
)

type (
	// DocumentMetaData is all of the data aside form tweets describing an entity
	// in the datastore.
	DocumentMetaData struct {
		Username                             string
		UserID, LastTweetID, EarliestTweetID int64
	}

	// CleanDocument is a collection of tweets (just the text) and some metadata
	// about the user and the collection itself.
	CleanDocument struct {
		DocumentMetaData
		Tweets []string
	}

	// AnalysedDocument is the actual entity that is stored in the datastore, all of
	// the tweets have been converted to scores. There is no way to get the original
	// tweet back from the score at this point.
	//AnalysedDocument struct {
	//	DocumentMetaData
	//	TweetScores []int
	//}
	AnalysedDocument struct {
		DocumentMetaData
		TweetScores                    []int
		PositiveTweets, NegativeTweets int
		AverageScore                   float64
	}

	// Changes is the flag in the database indicating whether new data has been
	// added since the last time Changes was checked and toggled.
	Changes struct{ ChangesMade bool }
)

const (
	changesKind = "Changes"
	userKind    = "User"
)

// changesKey is the key to an entity that indicates whether any updates to the database have happened
var changesKey = datastore.IDKey(changesKind, 0, nil)

// analyse performs analysis on all of the tweets in an object in cloud storage
// using the model, then returns all of the new data.
func analyse(obj *storage.ObjectHandle, model sentiment.Models) *AnalysedDocument {
	reader, err := obj.NewReader(context.Background())
	if err != nil {
		log.Println(err)
		return nil
	}
	if err := obj.Delete(context.Background()); err != nil {
		log.Println(err)
		return nil
	}
	decoder := json.NewDecoder(reader)
	doc := &CleanDocument{}
	if err := decoder.Decode(doc); err != nil {
		log.Println(err)
		return nil
	}
	newDoc := &AnalysedDocument{
		DocumentMetaData: DocumentMetaData{
			UserID:          doc.UserID,
			LastTweetID:     doc.LastTweetID,
			EarliestTweetID: doc.EarliestTweetID,
		},
		TweetScores: make([]int, len(doc.Tweets)),
	}
	positiveTweetCount := 0
	negativeTweetCount := 0
	totalSentimentScore := 0
	for i, tweet := range doc.Tweets {
		analysis := model.SentimentAnalysis(tweet, sentiment.English)
		newDoc.TweetScores[i] = int(analysis.Score)
		if analysis.Score == 0 {
			negativeTweetCount += 1
		} else if analysis.Score == 1 {
			positiveTweetCount += 1
		}
		totalSentimentScore += int(analysis.Score)
	}

	if len(newDoc.TweetScores) == 0 {
		return nil
	}
	newDoc.PositiveTweets = positiveTweetCount
	newDoc.NegativeTweets = negativeTweetCount
	newDoc.AverageScore = float64(totalSentimentScore) / float64(len(newDoc.TweetScores))
	//fmt.Println(doc.Tweets)
	return newDoc
}

// store takes in an AnalysedDocument and updates the necessary entities in
// the datastore.
//TODO: this might need changed
func store(doc *AnalysedDocument, ds *datastore.Client) error {
	key := datastore.IDKey(userKind, doc.UserID, nil)
	tx, err := ds.NewTransaction(context.Background())
	if err != nil {
		return err
	}
	// get the old entity
	oldDoc := &AnalysedDocument{}
	// TODO: Verify that we are making objects correctly, look up Update instead of Get->Put
	if err := tx.Get(key, oldDoc); err != nil {
		// pass, I think this means object is not in db yet, should be consistent
	} else if oldDoc.EarliestTweetID < doc.EarliestTweetID &&
		oldDoc.LastTweetID > doc.LastTweetID {
		// if all of the tweets in this batch have already been covered then do
		// not add them to the database as they have already been analysed.
		return tx.Rollback()
	} else {
		// combine new data and old data
		if oldDoc.EarliestTweetID < doc.EarliestTweetID {
			doc.EarliestTweetID = oldDoc.EarliestTweetID
		}
		if oldDoc.LastTweetID > doc.LastTweetID {
			doc.LastTweetID = oldDoc.LastTweetID
		}
		doc.TweetScores = append(doc.TweetScores, oldDoc.TweetScores...)
	}
	// put the entity in the db
	if _, err := tx.Put(key, doc); err != nil {
		return err
	}
	// commit the updates
	if _, err = tx.Commit(); err != nil {
		return err
	}
	// update the changes flag in the database
	_, err = ds.Put(context.Background(), changesKey, &Changes{ChangesMade: true})
	return err
}

// Analyse reads roughly cleaned tweets from the bucket and then performs
// sentiment analysis on them. After all of the analysis is complete, the new
// data is pushed to the datastore.
func Analyse(bucket *storage.BucketHandle, model sentiment.Models, ds *datastore.Client, fileName string) {
	doc := analyse(bucket.Object(fileName), model)
	if doc != nil {
		err := store(doc, ds)
		if err != nil {
			log.Println("error storing doc: ", err)
			return
		}
	}
	// while the quit signal has not been sent, keep trying to get more documents
	//for len(stop) == 0 {
	//	// Get all of the names of all objects in the bucket
	//	iter := bucket.Objects(context.Background(), &storage.Query{
	//		Versions:   false,
	//		Projection: storage.ProjectionNoACL,
	//	})
	//	names := make([]string, 0)
	//	for obj, err := iter.Next(); err == nil && obj != nil; obj, err = iter.Next() {
	//		names = append(names, obj.Name)
	//	}
	//	// If there are no names, then sleep for a bit then check again
	//	if len(names) == 0 {
	//		time.Sleep(10 * time.Second)
	//	} else {
	//		// Otherwise, read in the names one at a time and process the objects
	//		for _, name := range names {
	//			// If the quite signal has been sent before reading in any names
	//			// then stop analysing right now.
	//			if len(stop) > 0 {
	//				return
	//			}
	//			doc := analyse(bucket.Object(name), model)
	//			if doc != nil {
	//				store(doc, ds)
	//			}
	//		}
	//	}
	//}
}
