test:
	go test -v ./...

vet:
	go tool vet -v -cover ./

build: test
	CGO_ENABLED=0 go build -a --installsuffix cgo
