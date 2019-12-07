# Serve Server - Painless Docker Deployments

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