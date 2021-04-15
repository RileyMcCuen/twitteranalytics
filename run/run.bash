# Builds the go application, runs it, then opens a new tab in the browser
go build -o ../build/twitteranalytics .
open --new -a "Google Chrome" --args "localhost/static"
echo "Running..."
../build/twitteranalytics