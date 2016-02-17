SOURCES := $(filter-out $(wildcard *_test.go),$(wildcard *.go))

GO := $(shell command -v go)
GO_VERSION = 1.5.3 # for docker image only

TARGET := reaper

ifdef GO
	GO := GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 $(GO)
endif

ifndef GO
	GO := docker run --rm -v $(PWD):/go/src/github.com/trayio/reaper -w /go/src/github.com/trayio/reaper -e CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 golang:$(GO_VERSION) go
endif

$(TARGET): $(SOURCES)
	$(GO) build -a --installsuffix cgo -o $@

build: $(TARGET)

clean:
	rm -f $(TARGET)

test:
	$(GO) test -v ./...

.PHONY: build clean test
