lint:
	- golangci-lint run

build-linux:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o serve-server

up:
	docker-compose -p serve -f docker-compose.yml up -d --build serve-docker