all: style staticcheck build test

style:
  gofmt -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
  go test ./...

format:
  go fmt ./...

vet:
  go vet ./...

staticcheck:
  staticcheck ./...

build:
  go build

docker:


.PHONY: all style test format vet staticcheck build