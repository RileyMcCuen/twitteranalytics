# Builds the go application, runs it, then opens a new tab in the browser
$env:GOOGLE_APPLICATION_CREDENTIALS = "..\twitteranalytics-310723-f09ac30d22c2.json"
$env:BUCKET = "twittertimelines"
$env:PROJECT_ID = "twitteranalytics-310723"
go build -o ../build/nameindex.exe ..
../build/nameindex.exe
