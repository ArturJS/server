lint:
	- golangci-lint run

go-build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o serve-server

docker-build:
	docker build -t loeffel/serve -f Dockerfile .

docker-run:
	docker run -d \
      --restart always \
      --name serve \
      -p 8080:8080 \
      -v /var/run/docker.sock:/var/run/docker.sock:ro \
      -v ~/serve:/home/serve \
      -e MAX_SIZE=32 \
      -e TOKEN="Rdgn3-*CGkyNnB!f" \
      loeffel/serve

docker-watchtower:
	docker run -d \
    	--name watchtower \
    	-v /var/run/docker.sock:/var/run/docker.sock:ro \
    	v2tec/watchtower serve

build-run:
	make go-build
	make docker-build
	make docker-run
