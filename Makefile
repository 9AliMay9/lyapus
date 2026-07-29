.PHONY: fmt lint generate generate-check test race integration vuln run verify

fmt:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -n "$$files" ]; then gofmt -w $$files; fi

lint:
	go vet ./...

generate:
	@sqlc="$$(go env GOPATH)/bin/sqlc"; \
	if [ ! -x "$$sqlc" ]; then \
			echo "sqlc is not installed; run: go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1"; \
			exit 1; \
	fi; \
	if [ "$$($$sqlc version)" != "v1.31.1" ]; then \
			echo "sqlc version must be v1.31.1"; \
			exit 1; \
	fi; \
	"$$sqlc" generate

generate-check: generate
	@if [ -n "$$(git ls-files --others --exclude-standard -- internal/catalog/postgres/sqlcgen)" ] \
			|| ! git diff --quiet -- internal/catalog/postgres/sqlcgen; then \
			echo "generate sqlc code is stale or untracked; run: make generate and add the generated files"; \
			exit 1; \
	fi

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

verify: fmt lint generate-check test race integration vuln
