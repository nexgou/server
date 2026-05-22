# Nexgou Server — Makefile
# Usage:
#   make hooks      install git hooks (run once after cloning)
#   make test       run all tests with race detector
#   make coverage   run tests + open coverage report
#   make lint       run golangci-lint
#   make build      build all packages + samples
#   make vet        run go vet
#   make clean      remove generated artifacts
#   make ci         full CI pipeline (vet + lint + test + build)

.PHONY: hooks test coverage lint build vet clean ci

# ── Detect OS for hook installer ─────────────────────────────────────────────
ifeq ($(OS),Windows_NT)
INSTALL_HOOKS := powershell -ExecutionPolicy Bypass -File scripts/install-hooks.ps1
else
INSTALL_HOOKS := sh scripts/install-hooks.sh
endif

# ── Coverage threshold ────────────────────────────────────────────────────────
COVERAGE_THRESHOLD := 80

# ── Targets ───────────────────────────────────────────────────────────────────

## hooks: configure git to use .githooks/
hooks:
	$(INSTALL_HOOKS)

## vet: run go vet on all packages
vet:
	go vet ./...

## lint: run golangci-lint
lint:
	golangci-lint run --timeout=5m

## test: run tests with race detector
test:
	go test -v -race ./...

## coverage: run tests, generate coverage report, and check threshold
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@TOTAL=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo ""; \
	echo "Coverage: $${TOTAL}%  (threshold: $(COVERAGE_THRESHOLD)%)"; \
	if [ "$$(echo "$${TOTAL} < $(COVERAGE_THRESHOLD)" | bc -l)" = "1" ]; then \
		echo "FAIL: coverage is below $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "PASS: coverage meets the threshold"; \
	fi

## coverage-html: open coverage report in browser
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Report written to coverage.html"

## build: build all packages and sample binaries
build:
	go build ./...

## clean: remove generated artifacts
clean:
	rm -f coverage.out coverage.html
	go clean ./...

## ci: full pipeline (matches GitHub Actions)
ci: vet lint coverage build
	@echo ""
	@echo "CI pipeline passed."
