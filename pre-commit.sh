#!/bin/bash
# Copy this to .git/hooks/pre-commit (in the directory where the repo is)
# and make it executable with chmod +x .git/hooks/pre-commit
set -e

echo "Downloading test suites"
go get -v github.com/golang/lint/golint
go get github.com/gordonklaus/ineffassign
go get -u github.com/client9/misspell/cmd/misspell

#echo "Formatting with gofmt"
#gofmt -s -w ./*.go

echo "Running tests"
echo "-------------"
echo "go fmt test..."
diff <(gofmt -d .) <(echo -n)
echo "go fmt pass"
echo "vet test..."
go vet -x ./...
echo "vet pass"
echo "golint test..."
golint -set_exit_status ./...
echo "golint pass"
echo "ineffassign test..."
ineffassign ./
echo "ineffassign pass"
echo "misspell test..."
misspell . -error
echo "misspell pass"
echo "go tests..."
go test -v -race ./...
echo "go tests pass"
echo "Tests complete"