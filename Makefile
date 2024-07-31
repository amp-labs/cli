.PHONY: fix
fix:
	wsl --allow-cuddle-declarations --fix ./... && \
		gci write . && \
		golangci-lint run -c .golangci.yml --fix

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort
