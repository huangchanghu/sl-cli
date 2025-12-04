BINARY_NAME=sl-cli
INSTALL_PATH=/usr/local/bin

build:
	go build -o $(BINARY_NAME) cmd/sl-cli/main.go

install: build
	@echo "Installing to $(INSTALL_PATH)..."
	mv $(BINARY_NAME) $(INSTALL_PATH)
	@echo "Done! You can now run '$(BINARY_NAME)' anywhere."

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test ./...
