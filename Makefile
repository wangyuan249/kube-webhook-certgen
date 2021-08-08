# Repo info
GIT_COMMIT  ?= git-$(shell git rev-parse --short HEAD)

# Image URL to use all building/pushing image targets
CERT_GEN_IMAGE  ?= oamdev/kube-webhook-certgen:v2.2

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

build: reviewable
	go build -o main ./main.go

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

reviewable: fmt vet
	go mod tidy

# Run tests
test: vet
	go test ./pkg/...

docker-build:
	docker build -t $(CERT_GEN_IMAGE) .

docker-push:
	docker push  $(CERT_GEN_IMAGE)


