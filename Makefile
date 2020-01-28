linter:
	- golangci-lint run

go-build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o makeless-server

docker-build:
	docker build -t makeless/server -f Dockerfile .

docker-run:
	docker run -d \
      --restart always \
      --name makeless \
      -p 8080:8080 \
      -v /var/run/docker.sock:/var/run/docker.sock:ro \
      -v ~/makeless:/home/makeless \
      -v ~/certs:/home/certs \
      -e MAX_SIZE=32 \
      -e TOKEN="RANDOM-TOKEN-HERE" \
      makeless/server

docker-watchtower:
	docker run -d \
    	--name watchtower \
    	-v /var/run/docker.sock:/var/run/docker.sock:ro \
    	v2tec/watchtower makeless

build-run:
	make go-build
	make docker-build
	make docker-run
