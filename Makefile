.PHONY: fmt lint test race integration vuln run verify

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

vuln:
	@scanner="$$(go env GOPATH)/bin/govulncheck"; \
	if [ ! -x "$$scanner" ]; then \
		echo "govulncheck is not installed; run: go install golang.org/x/vuln/cmd/govulncheck@v1.6.0"; \
		exit 1; \
	fi; \
	"$$scanner" ./...

run:
	go run ./cmd/apiserver

verify: fmt lint test race integration vuln
