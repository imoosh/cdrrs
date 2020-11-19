# https://sohlich.github.io/post/go_makefile/


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

ifeq ($(shell uname), Linux)
all: linux
else ifeq ($(shell uname), Darwin)
all: darwin
else
endif

# development environment
dev: darwin

darwin :
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build $(GOBUILD_FLAGS)  -o $(BUILD_PATH)/bin/packet-analytic-service main/packet-analytic-service.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build $(GOBUILD_FLAGS)  -o $(BUILD_PATH)/bin/packet-input-service main/packet-input-service.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build $(GOBUILD_FLAGS)  -o $(BUILD_PATH)/bin/cdr-output-service main/cdr-output-service.go

# Cross compilation
linux :
	CGO_ENABLED=1 GOOS=linux  GOARCH=amd64  go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/packet-analytic-service main/packet-analytic-service.go
	CGO_ENABLED=1 GOOS=linux  GOARCH=amd64  go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/packet-input-service main/packet-input-service.go
	CGO_ENABLED=1 GOOS=linux  GOARCH=amd64  go build $(GOBUILD_FLAGS) -o $(BUILD_PATH)/bin/cdr-output-service main/cdr-output-service.go

test:
	go test -v ./...

clean:
	rm -rf $(OUTPUT_PATH)

run:
	go build -o $(BINARY_NAME) -v ./...
	./$(BINARY_MAXOSX)

fmt:
	go fmt ./...

# deps:

install:
	install -d $(OUTPUT_PATH)/tmp
	install -d $(OUTPUT_PATH)/bin
	install -d $(OUTPUT_PATH)/conf
	install -m 0755 $(SCRIPTS_PATH)/vsctl.sh $(OUTPUT_PATH)/
	install -m 0755 $(SCRIPTS_PATH)/cdr-output-service.sh $(OUTPUT_PATH)/
	install -m 0755 $(SCRIPTS_PATH)/packet-analytic-service.sh $(OUTPUT_PATH)/
	install -m 0755 $(SCRIPTS_PATH)/packet-input-service.sh $(OUTPUT_PATH)/
	#install $(SCRIPTS_PATH)/mysql.sql $(OUTPUT_PATH)/
	install $(SCRIPTS_PATH)/config.toml $(OUTPUT_PATH)/conf/
	install $(SCRIPTS_PATH)/20201015.pcapng $(OUTPUT_PATH)/tmp/
	install -m 0755 $(SCRIPTS_PATH)/tcpreplay.sh $(OUTPUT_PATH)/tmp/

docker:
