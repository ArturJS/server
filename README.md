<p align="center">
  <a href="https://vuepress.vuejs.org/" target="_blank">
    <img width="180" src="https://raw.githubusercontent.com/loeffel-io/serve-server/master/serve-logo.png" alt="logo">
  </a>
</p>

<p align="center">
  <a href="https://travis-ci.com/loeffel-io/serve-server"><img src="https://travis-ci.com/loeffel-io/serve-server.svg?token=diwUYjrdo8kHiwiMCFuq&branch=master" alt="Build"></a>
</p>

<h2 align="center">Supporting VuePress</h2>

# WIP: Serve Server - Painless Docker Deployments

[![Build Status](https://travis-ci.com/loeffel-io/serve-server.svg?token=diwUYjrdo8kHiwiMCFuq&branch=master)](https://travis-ci.com/loeffel-io/serve-server)

## Installation

- Replace `TOKEN`

```bash
docker run -d \
    --restart always \
    --name serve \
    -p 8080:8080 \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    -v ~/serve:/home/serve \
    -e MAX_SIZE=32 \
    -e TOKEN="RANDOM-TOKEN-HERE" \
    loeffel/serve

docker run -d \
    --name watchtower \
    -v /var/run/docker.sock:/var/run/docker.sock:ro \
    v2tec/watchtower serve
```

## Client

[Serve Client](https://github.com/loeffel-io/serve)