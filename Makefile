.PHONY: build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o mariadb-operator -installsuffix cgo ./cmd/
