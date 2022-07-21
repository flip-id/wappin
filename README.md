# Wappin

[![Go Reference](https://pkg.go.dev/badge/github.com/flip-id/wappin.svg)](https://pkg.go.dev/github.com/flip-id/wappin)
[![Go Report Card](https://goreportcard.com/badge/github.com/flip-id/wappin)](https://goreportcard.com/report/github.com/flip-id/wappin)

Wappin API client library. https://wappin.id.

# How to Test

To run the integration tests, we need to do the following:
1. Make a new copy of `.env.example` to `.env` by running this command below:
```bash
cat .env.example > .env
```
2. Fill the new .env with the parameter that we already prepared.
3. Run the tests by running this command:
```bash
go test -v -race -tags=integration -covermode=atomic ./...
```