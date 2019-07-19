prepare:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep ensure -v

build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bin/cli -a -v -installsuffix cgo -ldflags '-s -w' main.go

run:
	go run main.go

