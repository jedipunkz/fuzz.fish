XDG_CONFIG_HOME ?= $(HOME)/.config
FISH_CONFIG_DIR ?= $(XDG_CONFIG_HOME)/fish
FUNCTIONS_DIR := $(FISH_CONFIG_DIR)/functions
CONFD_DIR := $(FISH_CONFIG_DIR)/conf.d
BIN_PATH := $(FUNCTIONS_DIR)/fuzz

.PHONY: build test lint install uninstall reinstall

build:
	go build -o fuzz ./cmd/fuzz

test:
	go test ./...

lint:
	go vet ./...

# Leave the same files on disk as `fisher install jedipunkz/fuzz.fish`:
# the shell integration in conf.d/ and the binary in functions/.
install:
	mkdir -p $(FUNCTIONS_DIR) $(CONFD_DIR)
	go build -o $(BIN_PATH) ./cmd/fuzz
	cp conf.d/fuzz.fish $(CONFD_DIR)/fuzz.fish

uninstall:
	rm -f $(BIN_PATH) $(CONFD_DIR)/fuzz.fish

reinstall: uninstall install
