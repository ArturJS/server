<img width="180" src="https://raw.githubusercontent.com/makeless/server/master/makeless-logo.png" alt="logo">

# Makeless Server - Painless Docker Deployments

[![Build Status](https://travis-ci.com/makeless/server.svg?branch=master)](https://travis-ci.com/makeless/server)

## Installation

- Replace `TOKEN`

```bash
docker run -d \
    --restart always \
    --name makeless \
    -p 8080:8080 \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    -v ~/makeless:/home/makeless \
    -e MAX_SIZE=32 \
    -e TOKEN="RANDOM-TOKEN-HERE" \
    makeless/server

docker run -d \
    --name watchtower \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    v2tec/watchtower makeless
```