.PHONY: fmt test verify

fmt:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -n "$$files" ]; then gofmt -w $$files; fi

test:
	go test ./...

verify: fmt test
	go vet ./...
