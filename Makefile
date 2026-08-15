.PHONY: format format-go format-web format-check lint lint-go lint-web test vet build help web-install

GO_DIRS := cmd internal migrations
GO_PACKAGES := $(shell go list ./... | grep -v '/web/node_modules/')

format: format-go format-web

format-go:
	gofmt -w $(GO_DIRS)

format-web: web-install
	cd web && npx prettier --write .

format-check: web-install
	@unformatted=$$(gofmt -l $(GO_DIRS)); \
	if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi
	cd web && npx prettier --check .

lint: lint-go lint-web

lint-go:
	go vet $(GO_PACKAGES)

lint-web: web-install
	cd web && npx eslint . --no-error-on-unmatched-pattern

test:
	go test $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

build:
	go build -o bin/storycode ./cmd/storycode

help:
	go run ./cmd/storycode --help

web-install:
	cd web && npm ci
