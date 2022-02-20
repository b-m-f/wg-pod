SHELL := /bin/bash

build:
	CGO_ENABLED=0 go build

build-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build
