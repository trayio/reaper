test:
	go test -v ./...

vet:
	go tool vet -v ./

build: test
	CGO_ENABLED=0 go build -a --installsuffix cgo
