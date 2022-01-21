VERSION := $(shell git describe --tags)
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'

ARCHITECTURES = amd64 arm64
BUILD_TARGETS = $(patsubst %, apt-s3_$(VERSION)_%, $(ARCHITECTURES))
PACKAGE_TARGETS = $(patsubst %, apt-s3_$(VERSION)_%.deb, $(ARCHITECTURES))

all: test $(PACKAGE_TARGETS)

$(BUILD_TARGETS): apt-s3_$(VERSION)_% : build-deps
	GOOS=linux GOARCH=$* go1.12.17 build $(LDFLAGS) -o $@

$(PACKAGE_TARGETS): apt-s3_$(VERSION)_%.deb : apt-s3_$(VERSION)_%
	cp apt-s3_$(VERSION)_$* apt-s3 # Workaround, nfpm does not support env vars in contents
	VERSION=$(VERSION) ARCH=$* nfpm pkg --target $@
	rm apt-s3

build-deps:
	go1.12.17 get ./...

clean:
	rm -f apt-s3_* apt-s3_*.deb

release: $(PACKAGE_TARGETS) tag
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is not set!)
endif
	$(eval URL := $(shell curl -sS -H "Authorization: token $$GITHUB_TOKEN" -H "Content-Type: application/json" -X POST -d '{"tag_name":"$(VERSION)","name":"v$(VERSION)"}' https://api.github.com/repos/zendesk/apt-s3/releases | awk -F\" /assets_url/'{sub(/api/, "uploads", $$4); print $$4 }'))
	$(foreach arch,$(ARCHITECTURES),\
		$(shell curl -sS -H "Authorization: token $$GITHUB_TOKEN" -H "Content-Type: application/octet-stream" -X POST --data-binary "@apt-s3_$(VERSION)_$(arch)" $(URL)?name=apt-s3_$(VERSION)_$(arch) >/dev/null)\
		$(shell curl -sS -H "Authorization: token $$GITHUB_TOKEN" -H "Content-Type: application/octet-stream" -X POST --data-binary "@apt-s3_$(VERSION)_$(arch).deb" $(URL)?name=apt-s3_$(VERSION)_$(arch).deb >/dev/null)\
	)

tag:
	git tag $(VERSION)
	git push --tags

test: build-deps
	go test -v ./...

.PHONY: all build-deps clean release tag test
