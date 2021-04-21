# Builds the go application, runs it, then opens a new tab in the browser
go build -o ../build/twitter .
echo "Running..."
../build/twitter