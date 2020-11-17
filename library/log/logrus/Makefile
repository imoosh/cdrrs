# https://sohlich.github.io/post/go_makefile/

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main
BINARY_LINUX32=$(BINARY_NAME).linux-i386
BINARY_LINUX64=$(BINARY_NAME).linux-amd64

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX32)
	rm -f $(BINARY_LINUX64)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	$(GOGET) github.com/markbates/goth
	$(GOGET) github.com/markbates/pop

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINARY_LINUX32) -v
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_LINUX64) -v
docker-build:
	docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_UNIX)" -v
