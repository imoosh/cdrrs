# https://sohlich.github.io/post/go_makefile/

BINARY_NAME=vs
BINARY_MAC=$(BINARY_NAME)_darwin_amd64
BINARY_LINUX32=$(BINARY_NAME)_linux_386
BINARY_LINUX64=$(BINARY_NAME)_linux_amd64
BINARY_WINDOWS=$(BINARY_NAME)_windows_amd64
TOP_PATH=.
BUILD_PATH=$(TOP_PATH)/output
SCRIPTS_PATH=$(TOP_PATH)/scripts
OUTPUT_PATH=$(TOP_PATH)/output

BUILT_ID_TAG := main.BuiltID=$(shell git symbolic-ref --short HEAD 2>/dev/null) +$(shell git rev-parse --short HEAD)
BUILT_HOST_TAG := main.BuiltHost=$(shell whoami)@$(shell hostname)
BUILT_TIME_TAG := main.BuiltTime=$(shell date)
BUILT_GOVER_TAG := main.GoVersion=$(shell go version)

GOBUILD_FLAGS := -ldflags "-X \"$(BUILT_ID_TAG)\" -X \"$(BUILT_TIME_TAG)\" -X \"$(BUILT_HOST_TAG)\" -X \"$(BUILT_GOVER_TAG)\""

.PHONY : all build darwin linux32 linux64 win32 win64 test clean fmt install docker

all: darwin linux64

# development environment
dev: darwin

darwin :
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64  go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/$(BINARY_MAC) -v

# Cross compilation
linux32 :
	CGO_ENABLED=0 GOOS=linux  GOARCH=386    go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/$(BINARY_LINUX32) -v

linux64 :
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64  go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/$(BINARY_LINUX64) -v

win32 :
	CGO_ENABLED=0 GOOS=windows GOARCH=386   go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/$(BINARY_WINDOWS) -v

win64 :
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/$(BINARY_WINDOWS) -v

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_PATH)
	rm -rf $(OUTPUT_PATH)

run:
	go build -o $(BINARY_NAME) -v ./...
	./$(BINARY_MAXOSX)

fmt:
	go fmt ./...

# deps:

install:
	install -d $(OUTPUT_PATH)/bin
	install -d $(OUTPUT_PATH)/conf
	install -d $(OUTPUT_PATH)/audio
	install -m 0755 $(SCRIPTS_PATH)/asrctl.sh $(OUTPUT_PATH)/
	install -m 0755 $(BUILD_PATH)/bin/* $(OUTPUT_PATH)/bin/
	install $(SCRIPTS_PATH)/mysql.sql $(OUTPUT_PATH)/
	install $(SCRIPTS_PATH)/config.yml $(OUTPUT_PATH)/conf/
	install $(SCRIPTS_PATH)/config.toml $(OUTPUT_PATH)/conf/

docker:
