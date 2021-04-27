# This script should be run from inside of the build directory
$env:GOOS = "linux"
$env:GOARCH = "amd64"
# Go to root repository dir
Set-Location ..
# Build webserver
Set-Location .\webserver
go build -o .\build\twitteranalytics .
Set-Location ..
# Build twitter data fetcher
Set-Location .\twitter
go build -o .\build\twitter .
Set-Location ..
# Build data analyser
Set-Location .\analysis
go build -o .\build\analyse .
Set-Location ..
# Build name indexer
Set-Location .\name-index
go build -o .\build\nameindex .
Set-Location ..
# Return to build dir
Set-Location ./build