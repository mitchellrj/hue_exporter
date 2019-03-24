PREFIX                  ?= $(shell pwd)
VERSION                 ?= $(shell cat VERSION)

all: style staticcheck build test

style:
	! gofmt -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
	go test ./...

format:
	go fmt ./...

vet:
	go vet ./...

staticcheck:
	staticcheck ./...

promu:
	GOOS= GOARCH= go get -u github.com/prometheus/promu

build: promu
	promu build hue_exporter --prefix $(PREFIX)

crossbuild:
	promu crossbuild hue_exporter

docker: crossbuild
	docker build --pull -f Dockerfile.amd64 -t mitchellrj/hue_exporter:latest .
	docker tag mitchellrj/hue_exporter:latest mitchellrj/hue_exporter:$(VERSION)
	docker build --pull -f Dockerfile.arm7 -t mitchellrj/hue_exporter:latest-arm7 .
	docker tag mitchellrj/hue_exporter:latest-arm7 mitchellrj/hue_exporter:$(VERSION)-arm7

dist: docker

push:
	docker push mitchellrj/hue_exporter:latest
	docker push mitchellrj/hue_exporter:$(VERSION)
	docker push mitchellrj/hue_exporter:latest-arm7
	docker push mitchellrj/hue_exporter:$(VERSION)-arm7

DEFAULT: all
.PHONY: all style test format vet staticcheck promu build crossbuild dist docker push
