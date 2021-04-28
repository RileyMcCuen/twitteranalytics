# Builds the go application, runs it, then opens a new tab in the browser
$env:ACCESS_TOKEN = ""
$env:ACCESS_TOKEN_SECRET_KEY = ""
$env:API_KEY = ""
$env:API_SECRET_KEY = ""
$env:GOOGLE_APPLICATION_CREDENTIALS = ""
$env:BUCKET = "twittertimelines"
$env:PUB_SUB_SUBSCRIPTION_ID = "twitter-fetch"
$env:PUB_SUB_PUBLISH_ID = "twitter-documents"
$env:ADDRESS = "0.0.0.0:8001"
go build -o ../build/twitter.exe ..
Write-Output "Running..."
../build/twitter.exe