#!/bin/bash

# Generate protobuf files for Go services
echo "Generating protobuf files..."

# Create output directories
mkdir -p orchestrator-service/pkg/proto
mkdir -p agent-service/pkg/proto

# Generate for orchestrator service
protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/node_management.proto

echo "Protobuf generation completed!"