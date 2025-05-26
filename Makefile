.PHONY: build run-server run-client test clean

BUILD_DIR := bin
SERVER_BIN := $(BUILD_DIR)/server
CLIENT_BIN := $(BUILD_DIR)/client

build: $(SERVER_BIN) $(CLIENT_BIN)

$(SERVER_BIN): cmd/server/main.go pkg/server/*.go pkg/storage/*.go pkg/jenkins/*.go
	@mkdir -p $(BUILD_DIR)
	go build -o $@ cmd/server/main.go

$(CLIENT_BIN): cmd/client/main.go pkg/client/*.go
	@mkdir -p $(BUILD_DIR)
	go build -o $@ cmd/client/main.go

run-server: build
	@echo "Starting Log Proxy Server on port 8080..."
	./$(SERVER_BIN)

run-client: build
	@echo "Running Log Proxy Client..."
	./$(CLIENT_BIN) $(ARGS)
	## make run-client ARGS="-head 7372"
	## make run-client ARGS="7372 500 1000"
	## make run-client ARGS="7372 500"


test:
	@echo "Running all Go tests..."
	go test -v ./...

clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...