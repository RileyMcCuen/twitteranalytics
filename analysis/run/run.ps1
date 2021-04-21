# Builds the go application, runs it, then opens a new tab in the browser
$env:ACCESS_TOKEN = ""
$env:ACCESS_TOKEN_SECRET_KEY = ""
$env:API_KEY = ""
$env:API_SECRET_KEY = ""
$env:GOOGLE_APPLICATION_CREDENTIALS = ""
$env:ADDRESS = "0.0.0.0:80"
go build -o ../build/twitteranalytics.exe .
Start-Process -FilePath Chrome -ArgumentList localhost/static
Write-Output "Running..."
../build/twitteranalytics.exe
