package main

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/cdipaolo/sentiment"
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

	// AnalysedDocument is the actual entity that is stored in the datastore. It
	// contains metadata about the user and metrics about their tweets sentiment.
	AnalysedDocument struct {
		DocumentMetaData
		PositiveTweets, NegativeTweets int
		AverageScore                   float64
	}

	// Changes is the flag in the database indicating whether new data has been
	// added since the last time Changes was checked and toggled.
	Changes struct{ ChangesMade bool }
)

// CalculateAverage calculates the average sentiment of a document.
func (d *AnalysedDocument) CalculateAverage() *AnalysedDocument {
	score := float64((-1 * d.NegativeTweets) + d.PositiveTweets)
	count := float64(d.NegativeTweets + d.PositiveTweets)
	d.AverageScore = score / count
	return d
}

// Merge merges o's information into d. d is modified in the process.
func (d *AnalysedDocument) Merge(o *AnalysedDocument) *AnalysedDocument {
	// combine new data and old data
	if o.EarliestTweetID < d.EarliestTweetID {
		d.EarliestTweetID = o.EarliestTweetID
	}
	if o.LastTweetID > d.LastTweetID {
		d.LastTweetID = o.LastTweetID
	}
	d.NegativeTweets += o.NegativeTweets
	d.PositiveTweets += o.PositiveTweets
	return d.CalculateAverage()
}

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
	}
	for _, tweet := range doc.Tweets {
		analysis := model.SentimentAnalysis(tweet, sentiment.English)
		if analysis.Score == 0 {
			newDoc.NegativeTweets += 1
		} else if analysis.Score == 1 {
			newDoc.PositiveTweets += 1
		}
	}
	newDoc.AverageScore = float64((-1*newDoc.NegativeTweets)+newDoc.PositiveTweets) / float64(newDoc.NegativeTweets+newDoc.PositiveTweets)
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
		doc = doc.Merge(oldDoc)
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
}
