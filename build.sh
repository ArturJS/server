#!/bin/bash
set -e

GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -o serve-server
make up