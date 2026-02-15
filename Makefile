.PHONY: build install

build:
	go build ./cmd/fuzz

install:
	go build -o ~/.config/fish/functions/fuzz ./cmd/fuzz
