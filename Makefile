.PHONY: fix
fix:
	wsl --allow-cuddle-declarations --fix ./... && \
		gci write . && \
		golangci-lint run -c .golangci.yml --fix

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort

# Fix specific files passed as arguments
# Usage: make fix-files FILES="cmd/listen.go cmd/trigger.go"
fix-files:
	@if [ -z "$(FILES)" ]; then \
		echo "Usage: make fix-files FILES=\"file1.go file2.go ...\""; \
		echo "Example: make fix-files FILES=\"cmd/listen.go cmd/trigger.go\""; \
		exit 1; \
	fi
	@echo "Fixing files: $(FILES)"
	gci write $(FILES)
	@for file in $(FILES); do \
		echo "Formatting $$file..."; \
		gofmt -w $$file; \
	done
	@echo "Running go vet on packages containing the files..."
	@for file in $(FILES); do \
		dir=$$(dirname $$file); \
		echo "Vetting package ./$$dir/..."; \
		go vet ./$$dir/... || true; \
	done

# TODO: Add hot reloading for dev
# Build a CLI configured for the local stage
.PHONY: build/local
build/local:
	task build-local

# Build a CLI configured for the dev stage
.PHONY: build/dev
build/dev:
	task build-dev

# Build a CLI configured for the staging stage
.PHONY: build/staging
build/staging:
	task build-staging

# Build a CLI configured for the prod stage
.PHONY: build/prod
build/prod:
	task build-prod

# An alias for build-dev
.PHONY: build
build:
	task build