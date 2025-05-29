.PHONY: proto server client clean

# Generate protobuf and gRPC code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/greeter.proto

# Run the server
server:
	go run server/main.go

# Run the client
client:
	go run client/main.go

# Clean generated files
clean:
	rm -f pb/*.go
