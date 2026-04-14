.PHONY: help lint test build ci release-snapshot tidy

GO ?= go
GOLANGCI_LINT ?= golangci-lint
GORELEASER ?= goreleaser

help:
	@echo "Available targets:"
	@echo "  lint              Run golangci-lint (as in CI lint job)"
	@echo "  test              Run tests with race detector (as in CI test job)"
	@echo "  build             Build CLI binary package (as in CI build job)"
	@echo "  ci                Run lint + test + build locally"
	@echo "  tidy              Run go mod tidy"
	@echo "  release-snapshot  Run goreleaser snapshot build"

lint:
	$(GOLANGCI_LINT) run

test:
	$(GO) test -race -count=1 ./...

build:
	$(GO) build -v ./cmd/git-html-report/...

ci: lint test build

tidy:
	$(GO) mod tidy

release-snapshot:
	$(GORELEASER) release --snapshot --clean
