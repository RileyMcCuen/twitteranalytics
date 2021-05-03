# Twitter Feed Analysis

This project exists to analyse the sentiment and content of information being posted on a Twitter feed.

## Twitter API

We use the Twitter API to collect data for users in the form of tweets. This API is easy to use and credentials were very easy to get.

## Sentiment Analysis

Originally we used the Google Natural Language API to analyse various aspects of tweets that were received from
Twitter. However we soon realized it would be cost prohibitive to use the service because according to the pricing documentation it would cost approximately $3 to analyse a single highly active Twitter user.

Ultimately, we opted to import and use a popular open source library that performs sentiment analysis locally in Go, instead of using a large web API. The results will probably be lower quality, however, it is much cheaper to perfom analysis this way and with auto horizontal scaling this service should be able to handle all the load it needs to.

## Go REST API

A Go REST API is used to serve both static webpages and related content as well as dynamic content from the database. It also accepts requests that will eventually be passed off to another service to fetch data from Twitter and eventually analyse it.

## Kubernetes on GCP using Google Kubernetes Engine

All of the services for this application are run on Google Kubernetes Engine on GCP. A LoadBalancer service is used instead of an ingress that was used during local testing.

## Images hosted on Dockerhub

All of the Docker images are hosted on Dockerhub. All of their source code is in a single repository in Github. There is a special branch "feature/linux-builds" that contains amd64 linux binaries for all of the images. Dockerhub looks for any pushes to this branch of the repository and if it sees any, then it rebuilds all of the images. This process takes ~10 minutes and builds 4 images.

The images are:

- https://hub.docker.com/r/blunderingpb/twitter-analytics
    - This is the only entry point to project. It is the webserver that can request analysis, get a list of analysed entites, and serve static content.
- https://hub.docker.com/r/blunderingpb/twitter-fetcher
    - This component listens on a pub sub for messages, if it gets one then it will collect a bunch of tweets for the user and create a document in CloudStorage. Next it submits
      a new message to another pub sub to let the analyser know that there is data to be processed.
- https://hub.docker.com/r/blunderingpb/twitter-analyser
    - This component listens on a pub sub for messages, if it gets one then it will use the message to download a file form CloudStorage. It then reads in the document which 
      contains a list of tweets and user meta data. All of these tweets are passed into the analyser, collected, then added to the database. A flag in the database is set to           true to let the indexer know that people have been added.
- https://hub.docker.com/r/blunderingpb/twitter-indexer
    - This component runs on a cron schedule (every 30 minutes) in the cluster. It reads in a database flag, if it is true then it creates an index of all of the user names in 
      the database and stores it in a json file in CloudStorage. If the flag is false then it ends without performing any work.
## Kubernetes

Kubernetes is used to orchestrate the containers and allow easy deployment to the cloud.

The Kubernetes cluster is run on Google Kubernetes Engine on GCP. A LoadBalancer service is used instead of an ingress that was used during local testing, the ingress is still in the repository though.

This was a helpful link for creating volume to hold google credentials file: https://stackoverflow.com/questions/47021469/how-to-set-google-application-credentials-on-gke-running-through-kubernetes

## Frontend

There is a simple vanilla HTML, JS, and CSS frontend used to load data and allow the user to make requests.

The domain is hosted on Godaddy which points to the Kubernetes cluster. Here is a url for the site that is running: http://twitter-analytics-cse427.info/static/
NOTE: This site was made on 5-1-2021, realistically it will probably go offline in approximately the next month, but all of the code to run it will live here.
