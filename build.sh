#!/bin/bash
set -e
GOOS=linux GOARCH=amd64 go build -o serve-server
chmod +x serve-server
cp serve-server /Users/lucas.loeffel/go/src/github.com/loeffel-io/serve-docker/serve-server
(cd /Users/lucas.loeffel/go/src/github.com/loeffel-io/serve-docker && make up)