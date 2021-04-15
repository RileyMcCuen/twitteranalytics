# Builds the go application, runs it, then opens a new tab in the browser
go build -o ../build/twitteranalytics.exe .
Start-Process -FilePath Chrome -ArgumentList localhost/static
Write-Output "Running..."
../build/twitteranalytics.exe
