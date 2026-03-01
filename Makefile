VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w \
           -X github.com/ryanwersal/nepenthe/cmd.version=$(VERSION) \
           -X github.com/ryanwersal/nepenthe/cmd.commit=$(COMMIT) \
           -X github.com/ryanwersal/nepenthe/cmd.date=$(DATE)

.PHONY: build test lint clean install

build:
	@mkdir -p dist
	go build -ldflags '$(LDFLAGS)' -o dist/nepenthe .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf dist

install: build
	cp dist/nepenthe /usr/local/bin/nepenthe
