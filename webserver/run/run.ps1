# Builds the go application, runs it, then opens a new tab in the browser
# Twitter env variables
$env:ACCESS_TOKEN = "1094720198433783811-g3SeuryIvlgyXtnX84i1SDn5TuiYtK"
$env:ACCESS_TOKEN_SECRET_KEY = "Z7i0pLVs0AKeV0yChekuFQblBgmoNMthqsACqrb6xGISc"
$env:API_KEY = "RgiKQBtQQRhV0d0vngYfGRuR9"
$env:API_SECRET_KEY = "r2OwDZQKy7TX63YoiWqmojTnEUxVvnzgH1YjwnSCzSEPUtpZij"
# Google env variables
$env:GOOGLE_APPLICATION_CREDENTIALS = "../twitteranalytics-310723-f09ac30d22c2.json"
$env:BUCKET = "twittertimelines"
$env:PROJECT_ID = "twitteranalytics-310723"
$env:PUB_SUB_TOPIC_ID = "twitter-fetch"
# App config env variables
$env:ADDRESS = "0.0.0.0:80"
# Build the app
go build -o ../build/twitteranalytics.exe ..
# Open a tab in the browser pointed at the webserver
Start-Process -FilePath "Chrome" -ArgumentList "localhost/static/"
# Run the app
Write-Output "Running..."
../build/twitteranalytics.exe
