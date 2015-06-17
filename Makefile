test:
	go test -cover ./...

build: test
	CGO_ENABLED=0 go build -a --installsuffix cgo
