#!/bin/bash
set -e
GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o serve-server
make up