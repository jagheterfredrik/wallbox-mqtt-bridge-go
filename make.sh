set -x

CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o bridge-armhf .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bridge-arm64 .
