VERSION := $(shell git describe --tags)
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'

all: apt-s3_$(VERSION)_amd64.deb

clean:
	rm -f apt-s3 apt-s3_$(VERSION)_amd64.deb

apt-s3:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $@

apt-s3_$(VERSION)_amd64.deb: apt-s3
	VERSION=$(VERSION) nfpm pkg --target $@

.PHONY: clean all
