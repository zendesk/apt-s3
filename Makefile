VERSION := $(shell git describe --tags)
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'

all: test

apt-s3: build-deps
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $@

apt-s3_$(VERSION)_amd64.deb: apt-s3
	VERSION=$(VERSION) nfpm pkg --target $@

build-deps:
	go get ./...

clean:
	rm -f apt-s3 apt-s3_*_amd64.deb

test: apt-s3_$(VERSION)_amd64.deb
	go test -v ./...

.PHONY: all build-deps clean test
