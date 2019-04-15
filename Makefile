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

release: tag
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is not set!)
endif
	$(eval URL := $(shell curl -sS -H "Authorization: token $$GITHUB_TOKEN" -H "Content-Type: application/json" -X POST -d '{"tag_name":"$(VERSION)","name":"v$(VERSION)"}' https://api.github.com/repos/zendesk/apt-s3/releases | awk -F\" /assets_url/'{sub(/api/, "uploads", $$4); print $$4 }'))
	curl -sS -H "Authorization: token $$GITHUB_TOKEN" -X POST -F "data=@apt-s3" $(URL)?name=apt-s3 >/dev/null
	curl -sS -H "Authorization: token $$GITHUB_TOKEN" -X POST -F "data=@apt-s3_$(VERSION)_amd64.deb" $(URL)?name=apt-s3_$(VERSION)_amd64.deb >/dev/null

tag: apt-s3_$(VERSION)_amd64.deb
	git tag $(VERSION)
	git push --tags

test: apt-s3_$(VERSION)_amd64.deb
	go test -v ./...

.PHONY: all build-deps clean release tag test
