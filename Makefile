.PHONY: swagger build test

SWAG ?= ./.bin/swag

.bin/swag:
	mkdir -p .bin
	GOBIN="$(CURDIR)/.bin" GOSUMDB=off go install github.com/swaggo/swag/cmd/swag@v1.16.4

swagger: .bin/swag
	$(SWAG) init -g cmd/server/main.go -o docs --parseDependency --parseInternal

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...
