BINARY_NAME=margi
CMD_PATH=./cmd/margi
DIST_PATH=./dist
INSTALL_PATH=$(HOME)/.local/bin

build:
	go build -o $(DIST_PATH)/$(BINARY_NAME) $(CMD_PATH)

install:
	go build -o $(INSTALL_PATH)/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
