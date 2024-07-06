.PHONY: test test-coverage proto

# make command to run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# make command to run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# make command to compile protobuf files
proto:
	@echo "Compiling protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
