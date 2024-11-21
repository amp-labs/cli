.PHONY: fix
fix:
	wsl --allow-cuddle-declarations --fix ./... && \
		gci write . && \
		golangci-lint run -c .golangci.yml --fix

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort

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