# This script should be run from inside of the build directory
GOOS="linux"
GOARCH="amd64"
# Go to root repository dir
cd ..
# Build webserver
cd ./webserver
go build -o ./build/twitteranalytics .
cd ..
# Build twitter data fetcher
cd ./twitter
go build -o ./build/twitter .
cd ..
# Build data analyser
cd ./analysis
go build -o ./build/analyse .
cd ..
# Build name indexer
cd ./name-index
go build -o ./build/name-index .
cd ..
# Return to build dir
cd ./build
