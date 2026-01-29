
.DEFAULT_GOAL := test

# Get version from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
BIN_DIR := bin
DIST_DIR := dist
PLATFORMS := linux/amd64 darwin/amd64 darwin/arm64 windows/amd64

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run build command
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BIN_DIR)/kiki .

.PHONY: build/all
build/all:
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		OUT="$(DIST_DIR)/kiki_$${OS}_$${ARCH}"; \
		if [ "$$OS" = "windows" ]; then OUT="$$OUT.exe"; fi; \
		echo "Building $$OS/$$ARCH -> $$OUT"; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o $$OUT .; \
	done


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

.PHONY: package
package: build/all
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		BIN="kiki_$${OS}_$${ARCH}"; \
		if [ "$$OS" = "windows" ]; then \
			zip -j "$(DIST_DIR)/kiki_$(VERSION)_$${OS}_$${ARCH}.zip" "$(DIST_DIR)/$${BIN}.exe"; \
		else \
			tar -C "$(DIST_DIR)" -czf "$(DIST_DIR)/kiki_$(VERSION)_$${OS}_$${ARCH}.tar.gz" "$${BIN}"; \
		fi; \
	done

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

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
