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
as well as deliver data to the frontend.

The API will be deployed as a Docker image to GKE.

## Kubernetes on GCP using Google Kubernetes Engine

Will scale the backend under load and make sure the service is always available.

## Frontend - ?

A frontend to display the data from the language processing. Will use some sort
of simple charting API to use the data generated on the backend to chart
sentiment and other natural language analysis outputs.
