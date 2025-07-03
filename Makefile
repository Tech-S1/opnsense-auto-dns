# Makefile for opnsense-auto-dns

BINARY=bin/opnsense-auto-dns

.PHONY: all build clean run tidy

all: build

build:
	@mkdir -p bin
	go build -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -rf bin

tidy:
	go mod tidy 