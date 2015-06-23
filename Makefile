test:
	go test ./...

build: test
	CGO_ENABLED=0 go build -a --installsuffix cgo
