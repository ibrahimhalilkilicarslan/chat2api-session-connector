GO ?= go
VERSION ?= 0.1.0

.PHONY: fmt vet test vuln build check release clean

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test -race -cover ./...

vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...

build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath \
		-ldflags "-s -w -X github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/version.Version=$(VERSION)" \
		-o bin/chat2api-connector ./cmd/chat2api-connector

check: fmt vet test vuln build
	git diff --check

release:
	VERSION=$(VERSION) ./scripts/build-release.sh

clean:
	rm -rf bin dist coverage.out
