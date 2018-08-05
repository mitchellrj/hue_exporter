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

build:
	go build

dist: darwin amd64 arm7 docker

darwin:
	cp hue_exporter build/hue_exporter.darwin

amd64:
	docker build --pull -f Dockerfile.build.amd64 -t hue_exporter_builder:latest .
	docker run -v $$(pwd)/build:/build hue_exporter_builder:latest

arm7:
	docker build --pull -f Dockerfile.build.arm7 -t hue_exporter_builder:latest-arm .
	docker run -v $$(pwd)/build:/build hue_exporter_builder:latest-arm

docker:
	docker build --pull -f Dockerfile.amd64 -t mitchellrj/hue_exporter:latest .
	docker tag mitchellrj/hue_exporter:latest mitchellrj/hue_exporter:$$(build/hue_exporter.darwin -V)
	docker build --pull -f Dockerfile.arm7 -t mitchellrj/hue_exporter:latest-arm7 .
	docker tag mitchellrj/hue_exporter:latest mitchellrj/hue_exporter:$$(build/hue_exporter.darwin -V)-arm7

push:
	docker push mitchellrj/hue_exporter:latest
	docker push mitchellrj/hue_exporter:$$(build/hue_exporter.darwin -V)
	docker push mitchellrj/hue_exporter:latest-arm7
	docker push mitchellrj/hue_exporter:$$(build/hue_exporter.darwin -V)-arm7

.PHONY: all style test format vet staticcheck build