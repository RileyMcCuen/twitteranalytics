# Builds the go application, runs it, then opens a new tab in the browser
$env:GOOGLE_APPLICATION_CREDENTIALS = ""
$env:BUCKET = "twittertimelines"
$env:PROJECT_ID = "twitteranalytics-310723"
$env:ADDRESS = "0.0.0.0:8002"
go build -o ../build/analyse.exe ..
Write-Output "Running..."
../build/analyse.exe