BINARY_NAME=margi
CMD_PATH=./cmd/margi
INSTALL_PATH=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) $(CMD_PATH)

install:
	go build -o $(INSTALL_PATH)/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
