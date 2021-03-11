OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
ALL_PLATFORM = linux/amd64

REGISTRY ?= appsent
IMG ?= scw-clerk
FULL_IMG ?= $(REGISTRY)/$(IMG)

all: fmt vet
	go build -o bin/scw-clerk ./cmd/

run: fmt vet
	go run ./cmd/main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Build the docker image
docker-build:
	docker build --platform=linux/$(ARCH) -f Dockerfile . -t ${FULL_IMG}
