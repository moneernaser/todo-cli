#!/bin/bash

set -e

echo "Building todo CLI..."
CGO_ENABLED=1 go build -ldflags="-s -w" -o todo .

echo "Installing to /usr/local/bin/ (requires sudo)..."
sudo mv todo /usr/local/bin/

echo "Done! You can now run 'todo' from anywhere."
