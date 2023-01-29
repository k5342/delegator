.PHONY: build test clean

.DEFAULT_GOAL := build

build: go.sum
	go build -o delegator

test: go.sum
	go test

clean:
	rm delegator
