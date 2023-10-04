# Define the name of your binary
BINARY_NAME=main

# Directory where your main.go file is located
CMD_DIR=cmd

build:
	go build -o bin/$(BINARY_NAME) $(CMD_DIR)/$(BINARY_NAME).go

run:
	go run $(CMD_DIR)/$(BINARY_NAME).go
