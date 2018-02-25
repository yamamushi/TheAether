#!/bin/bash

echo "Downloading test suites"
go get -v github.com/golang/lint/golint
go get github.com/gordonklaus/ineffassign
go get -u github.com/client9/misspell/cmd/misspell

echo "Formatting with gofmt"
gofmt -s -w ./*.go

echo "Running tests"
diff <(gofmt -d .) <(echo -n)
go vet -x ./...
golint -set_exit_status ./...
ineffassign ./
misspell . -error
go test -v -race ./...

echo "Tests complete"