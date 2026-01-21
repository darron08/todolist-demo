#!/bin/bash

# Build script for Todo List API

set -e

echo "Starting build process..."

# Set build variables
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
VERSION=${VERSION:-latest}

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -X main.GitBranch=${GIT_BRANCH}"

echo "Building version: ${VERSION}"
echo "Git commit: ${GIT_COMMIT}"
echo "Git branch: ${GIT_BRANCH}"
echo "Build time: ${BUILD_TIME}"

# Create bin directory if it doesn't exist
mkdir -p bin

# Build for current platform
echo "Building for current platform..."
go build -ldflags "${LDFLAGS}" -o bin/todolist-api cmd/api/main.go

# Build for Linux AMD64
echo "Building for Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/todolist-api-linux-amd64 cmd/api/main.go

# Build for Linux ARM64
echo "Building for Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/todolist-api-linux-arm64 cmd/api/main.go

# Build for Windows AMD64
echo "Building for Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/todolist-api-windows-amd64.exe cmd/api/main.go

# Build for Darwin AMD64
echo "Building for Darwin AMD64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/todolist-api-darwin-amd64 cmd/api/main.go

# Build for Darwin ARM64
echo "Building for Darwin ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/todolist-api-darwin-arm64 cmd/api/main.go

echo "Build completed successfully!"
echo "Binaries created in bin/ directory:"
ls -la bin/

echo "Generating checksums..."
cd bin
sha256sum * > checksums.txt
cd ..

echo "Build process completed!"