#!/usr/bin/env bash

echo building...

sudo rm /var/app/current/go.*

# Install dependencies.
go get ./...
# Build app
GOOS=linux GOARCH=amd64 go build -o application -ldflags="-s -w"
