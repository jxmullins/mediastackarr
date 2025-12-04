#!/bin/bash
# Test build script for MediaStack CLI
# Builds and tests the CLI in an isolated Docker container

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "MediaStack CLI - Isolated Build Test"
echo "========================================"
echo ""

# Build the Docker image
echo "Step 1: Building Go CLI in Docker container..."
echo ""

docker build -t mediastack-cli-builder --target builder .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Build successful!"
    echo ""

    # Show version
    echo "Step 2: Verifying build..."
    docker run --rm mediastack-cli-builder ./build/mediastack version
    echo ""

    # Show help
    echo "Step 3: Showing CLI help..."
    docker run --rm mediastack-cli-builder ./build/mediastack --help
    echo ""

    # Extract binary if requested
    if [ "$1" == "--extract" ]; then
        echo "Step 4: Extracting binary..."
        mkdir -p build
        docker run --rm -v "$SCRIPT_DIR/build:/out" mediastack-cli-builder cp ./build/mediastack /out/
        echo "✅ Binary extracted to: $SCRIPT_DIR/build/mediastack"
        echo ""
    fi

    echo "========================================"
    echo "Build test completed successfully!"
    echo "========================================"
    echo ""
    echo "To extract the binary, run:"
    echo "  ./test-build.sh --extract"
    echo ""
    echo "To build the minimal runtime image:"
    echo "  docker build -t mediastack-cli ."
    echo ""
    echo "To run the CLI from Docker:"
    echo "  docker run --rm mediastack-cli status"
    echo ""
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi
