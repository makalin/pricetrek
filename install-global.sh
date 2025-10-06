#!/bin/bash

# PriceTrek Global Installation Script
# This script installs PriceTrek globally so it can be run from any directory

set -e

echo "🚀 Installing PriceTrek globally..."

# Build the project first
echo "📦 Building PriceTrek..."
make build

# Determine installation directory
if [[ -w "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
    echo "📁 Installing to system directory: $INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    echo "📁 Installing to user directory: $INSTALL_DIR"
    
    # Create directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "🔧 Adding $INSTALL_DIR to PATH..."
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
        echo "✅ Added to ~/.zshrc. Please run 'source ~/.zshrc' or restart your terminal."
    fi
fi

# Copy binary
echo "📋 Copying binary to $INSTALL_DIR..."
cp build/pricetrek "$INSTALL_DIR/"

# Make executable
chmod +x "$INSTALL_DIR/pricetrek"

# Verify installation
echo "✅ Installation complete!"
echo "🔍 Verifying installation..."

if command -v pricetrek &> /dev/null; then
    echo "✅ PriceTrek is now available globally!"
    echo "📍 Location: $(which pricetrek)"
    echo "🔢 Version: $(pricetrek --version)"
    echo ""
    echo "🎉 You can now run 'pricetrek' from any directory!"
    echo "📚 Run 'pricetrek help' to see all available commands."
else
    echo "❌ Installation verification failed."
    echo "💡 Try running 'source ~/.zshrc' or restart your terminal."
    exit 1
fi