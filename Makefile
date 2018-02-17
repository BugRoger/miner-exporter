DATE     = $(shell date +%Y%m%d%H%M)
IMAGE    ?= bugroger/miner-exporter
VERSION  = v$(DATE)
GOOS     ?= $(shell go env | grep GOOS | cut -d'"' -f2)
BINARIES := miner-exporter

LDFLAGS := -X github.com/bugroger/miner-exporter/main.VERSION=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS)"

SRCDIRS  := .
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: all clean

all: $(BINARIES:%=bin/$(GOOS)/%)

bin/%: $(GOFILES) Makefile
	GOOS=$(*D) GOARCH=amd64 go build $(GOFLAGS) -v -i -o $(@D)/$(@F) . 

build: $(BINARIES:%=bin/linux/%)
	docker build $(BUILD_ARGS) -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

clean:
	rm -rf bin/*

