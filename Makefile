.PHONY: test vet help

test:
	go test ./...

vet:
	go vet ./...

help:
	go run ./cmd/storycode --help
