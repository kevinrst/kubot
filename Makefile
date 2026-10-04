BINARY := kubot
VERSION ?= dev

.PHONY: build test vet lint fmt clean tidy install schema

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/kubot

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w cmd internal tools

tidy:
	go mod tidy

schema:
	go run ./tools/schemagen

install: build
	install -m 0755 bin/$(BINARY) $(or $(KUBOT_INSTALL_DIR),/usr/local/bin)/$(BINARY)

clean:
	rm -rf bin
