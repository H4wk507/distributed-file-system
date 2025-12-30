.PHONY: proto proto-install clean build test

# Generate Go code from proto files
proto:
	@echo "Generating protobuf code..."
	@mkdir -p dfs/pb
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/node.proto

# Install protoc plugins (run once)
proto-install:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Clean generated files
clean:
	rm -rf dfs/pb/*.pb.go

# Build the project
build:
	go build -o bin/api ./cmd/api
	go build -o bin/node ./cmd/node

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod tidy
	go mod download

