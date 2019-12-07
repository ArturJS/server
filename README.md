# serve-server

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
```
