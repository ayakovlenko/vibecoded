#!/usr/bin/env bash

set -e

BINARY_NAME="update-repos"
INSTALL_DIR="${HOME}/.local/bin"

echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME" ./cmd/update-repos

echo "Creating install directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

echo "Installing $BINARY_NAME to $INSTALL_DIR"
mv "$BINARY_NAME" "$INSTALL_DIR/"

chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Installation complete!"
echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
	echo ""
	echo "WARNING: $INSTALL_DIR is not in your PATH"
	echo "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
	echo "export PATH=\"\$PATH:$INSTALL_DIR\""
	echo ""
	echo "Or run with full path: $INSTALL_DIR/$BINARY_NAME"
else
	echo "You can now run: $BINARY_NAME [directory]"
fi
