
.DEFAULT_GOAL := test

# Get version from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run build command
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: build
build:
	go build $(LDFLAGS) -o bin/kiki .


# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run test commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

# run go tests
.PHONY: test
test:
	go test ./...

# check for data race conditions
.PHONY: test/race
test/race:
	go test -race ./...

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run code format and code style commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

# run go vet tool
.PHONY: vet
vet:
	go vet ./...

# run staticcheck tool
.PHONY: staticcheck
staticcheck:
	staticcheck ./...

# run all tools
.PHONY: check
check: vet staticcheck

.PHONY: vuln
vuln:
	go tool govulncheck ./...

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run release commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: release/patch
release/patch:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v0.0.1"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
    		PATCH=$$(echo $$LATEST_TAG | cut -d. -f3); \
    		NEW_PATCH=$$((PATCH + 1)); \
    		NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"


.PHONY: release/minor
release/minor:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v0.1.0"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
    		NEW_MINOR=$$((MINOR + 1)); \
    		NEW_TAG="v$$MAJOR.$$NEW_MINOR.0"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"

.PHONY: release/major
release/major:
		@if [ $$(git tag | wc -l) -eq 0 ]; then \
    		NEW_TAG="v1.0.0"; \
    	else \
    		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
    		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
    		NEW_MAJOR=$$((MAJOR + 1)); \
    		NEW_TAG="v$$NEW_MAJOR.0.0"; \
    	fi; \
    	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
    	echo "Created new tag: $$NEW_TAG"