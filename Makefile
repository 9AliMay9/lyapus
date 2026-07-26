.PHONY: fmt lint test race integration run verify

fmt:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -n "$$files" ]; then gofmt -w $$files; fi

lint:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

integration:
	go test -tags=integration -count=1 ./...

run:
	go run ./cmd/apiserver

verify: fmt lint test race integration
