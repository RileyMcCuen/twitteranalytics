# Twitter Feed Analysis

This project exists to analyse the sentiment and content of information being
posted on a Twitter feed.

## Twitter API

Use the Twitter API to get a collection of tweets for a user.

## Google Natural Language API

Use the GNL API to analyse various aspects of tweets that were received from
Twitter.

Generate important anaylitics and return these to the front end.

## Golang RESTful API

This will handle all requests from the frontend and proxy data between the APIs,
as well as deliver data back to the frontend.

The API will be deployed as a Docker image to GKE.

## Kubernetes on GCP using Google Kubernetes Engine

## Images hosted on Dockerhub

All of the images are hosted on Dockerhub.
All of their source code is in a single repository in Github.
There is a special branch "feature/linux-builds" that contains amd64 linux
binaries for all services that will be used in the images. Dockerhub looks for
any pushes to this branch of the repository and if it sees any, then it rebuilds
all of the images. This process takes ~10 minutes and builds 4 images.

The images are:

-   https://hub.docker.com/r/blunderingpb/twitter-analytics
-   https://hub.docker.com/r/blunderingpb/twitter-fetcher
-   https://hub.docker.com/r/blunderingpb/twitter-analyser
-   https://hub.docker.com/r/blunderingpb/twitter-indexer

## Kubernetes

Will scale the backend under load and make sure the service is always available.

## Frontend - ?

A frontend to display the data from the language processing. Will use some sort
of simple charting API to use the data generated on the backend to chart
sentiment and other natural language analysis outputs.
