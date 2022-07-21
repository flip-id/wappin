.ONESHELL:
SHELL=/bin/sh

# Input Variables with Default Value
GO_VERSION?=1.17
GOOS?=linux
GOARM?=7
GOARCH?=amd64
CGO_ENABLED?=0
DOC_PORT?=6060

.PHONY: default
default: help
	@# Help: The default target of the Makefile.

.PHONY: help
help:
	@# Help: Show the help description.

	printf "%-20s %s\n" "Target" "Description"
	printf "%-20s %s\n" "------" "-----------"
	make -pqR : 2>/dev/null \
		| awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' \
		| sort \
		| egrep -v -e '^[^[:alnum:]]' -e '^$@$$' \
		| xargs -I _ sh -c 'printf "%-20s " _; make _ -nB | (grep -i "^# Help:" || echo "") | tail -1 | sed "s/^# Help: //g"'

.PHONY: mod-tidy
mod-tidy:
	@# Help: Tidy the Go dependencies.

	@echo go mod tidy -compat=${GO_VERSION} -v
	go mod tidy -compat=${GO_VERSION} -v

.PHONY: lint
lint:
	@# Help: Lint the Go code.

	go fmt .
	go fmt $(go list ./... | grep -v /vendor/)

.PHONY: generate
generate:
	@# Help: Run the Go generator.

	go install -v github.com/vektra/mockery/v2@latest
	go generate -v ./...

.PHONY: test
test: generate
	@# Help: Run the go test.

	go test -v -race ./...

.PHONY: test-coverage
test-coverage: generate
	@# Help: Run the test to get coverage.

	go test -race -covermode=atomic -coverprofile=profile.cov $(go list ./... | grep -v /vendor/)

.PHONY: coverage
coverage: test-coverage
	@# Help: Run all actions to extract coverage data from the cover profile.
	go tool cover -func profile.cov

	# last line will be number of total coverage
	go tool cover -func profile.cov | tail -n1 | awk '{print $$3}'

.PHONY: mod-download
mod-download:
	@# Help: Download all the depencies needed for the Golang service.

	GOOS=${GOOS} GOARM=${GOARM} GOARCH=${GOARCH} go mod download

.PHONY: doc
doc:
	@# Help: Show the package documentation in the local server.

	godoc -http=:${DOC_PORT}