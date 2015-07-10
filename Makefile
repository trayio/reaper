PROJECT = reaper
GOPATH := $(GOPATH)
USER = $(shell id -un)
IMAGE = golang:1.4.2

DOCKER := docker run --rm -v $(PWD):/go/src/github.com/trayio/$(PROJECT) -w /go/src/github.com/trayio/$(PROJECT) -v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro -u $(USER):$(USER) $(IMAGE)
DOCKER_BUILD_STATIC := docker run --rm -v $(PWD):/go/src/github.com/trayio/$(PROJECT) -w /go/src/github.com/trayio/$(PROJECT) -e CGO_ENABLED=0 -v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro -u $(USER):$(USER) $(IMAGE)


test:
	$(DOCKER) go test -v ./...

build:
	$(DOCKER_BUILD_STATIC) go build -a --installsuffix cgo -o $(PROJECT) .

clean:
	rm -f $(PROJECT)

.PHONY: test build clean
