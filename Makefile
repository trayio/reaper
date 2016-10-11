SOURCES := $(filter-out $(wildcard *_test.go),$(wildcard *.go))

GO_VERSION = 1.7 # for docker image only

TARGET := reaper

$(TARGET): $(SOURCES)
	docker run \
		--rm \
		-v $(PWD):/go/src/github.com/trayio/$(TARGET) \
		-w /go/src/github.com/trayio/$(TARGET) \
		-e CGO_ENABLED=0 \
		golang:$(GO_VERSION) go build --ldflags '-extldflags "-static"' -o $(TARGET)

test:
	docker run \
		--rm \
		-v $(PWD):/go/src/github.com/trayio/$(TARGET) \
		-w /go/src/github.com/trayio/$(TARGET) \
		golang:$(GO_VERSION) bash -c 'go test -race -v `go list ./... | grep -v vendor`'

build: $(TARGET)

clean:
	rm -f $(TARGET)

.PHONY: build clean test
