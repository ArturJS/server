# serve-server

```bash
docker run -d \
  --restart always \
  --name serve \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v ~/serve:/home/serve \
  loeffel/serve
```
