#!/bin/bash

set -e  # Exit on any error

# Configuration
LIMA_INSTANCE_NAME="nyx-builder-$(date +%s)"
IMAGE_NAME="nyx-htmx"
IMAGE_TAG="$1"
BUILD_DIR="$(pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Cleanup function
cleanup() {
    log "Cleaning up..."
    if limactl list | grep -q "$LIMA_INSTANCE_NAME"; then
        log "Stopping and deleting Lima instance: $LIMA_INSTANCE_NAME"
        limactl stop "$LIMA_INSTANCE_NAME" || true
        limactl delete "$LIMA_INSTANCE_NAME" || true
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT



# Local build steps
local_build() {
    log "Starting local build process..."
    
    # Create static/css directory if it doesn't exist
    mkdir -p static/css
    
    # Generate Tailwind CSS
    log "Generating Tailwind CSS..."
    if [ -f "src/app.css" ]; then
        ./tailwindcss -i ./src/app.css -o ./static/css/index.css --config ./src/tailwindcss.config.js --minify
        success "Tailwind CSS generated"
    else
        warn "src/app.css not found, skipping Tailwind CSS generation"
    fi
    
    # Generate templ files
    log "Generating templ files..."
    templ generate ./src/
    success "Templ files generated"
    
    # Verify Go module
    log "Verifying Go module..."
    go mod tidy
    success "Go module verified"
}

# Create Lima instance and build container
container_build() {
    log "Creating Lima instance: $LIMA_INSTANCE_NAME"
    
    # Start Lima instance with default configuration
    limactl start --name="$LIMA_INSTANCE_NAME" --vm-type=vz --rosetta
    
    # Wait for instance to be ready
    log "Waiting for Lima instance to be ready..."
    while ! limactl shell "$LIMA_INSTANCE_NAME" -- nerdctl --version &> /dev/null; do
        sleep 2
        log "Still waiting for nerdctl to be available..."
    done
    
    success "Lima instance is ready"
    
    # Check for Containerfile
    log "Looking for Containerfile..."
    if [ ! -f "Containerfile" ]; then
        error "No Containerfile found in current directory"
        exit 1
    fi
    
    success "Found Containerfile"
    
    # Build container image
    log "Building container image using nerdctl..."
    limactl shell "$LIMA_INSTANCE_NAME" -- nerdctl build -f Containerfile -t "${IMAGE_NAME}:${IMAGE_TAG}" .
    
    success "Container image built successfully: ${IMAGE_NAME}:${IMAGE_TAG}"
    
    # Create exports directory and save image
    log "Saving image to exports/images/..."
    mkdir -p exports/images
    
    # Save inside Lima instance first
    limactl shell "$LIMA_INSTANCE_NAME" -- nerdctl save "${IMAGE_NAME}:${IMAGE_TAG}" -o "/tmp/${IMAGE_NAME}-${IMAGE_TAG}.tar"
    
    # Copy the tar file out of Lima to local exports directory
    limactl copy "$LIMA_INSTANCE_NAME:/tmp/${IMAGE_NAME}-${IMAGE_TAG}.tar" "exports/images/${IMAGE_NAME}-${IMAGE_TAG}.tar"
    
    success "Image saved to exports/images/${IMAGE_NAME}-${IMAGE_TAG}.tar"
    
    # Show image info
    log "Image information:"
    limactl shell "$LIMA_INSTANCE_NAME" -- nerdctl images "${IMAGE_NAME}:${IMAGE_TAG}"
    
    # Clean up - no temporary files to remove
}

# Main execution
main() {
    # Check if version argument is provided
    if [ -z "$IMAGE_TAG" ]; then
        error "Must provide a version number"
        echo "Usage: $0 <version>"
        echo "Example: $0 v1.0.0"
        exit 1
    fi
    
    log "Starting Nyx application build process"
    log "Building image: ${IMAGE_NAME}:${IMAGE_TAG}"
    
    local_build
    container_build
    
    success "Build process completed successfully!"
    success "Container image: ${IMAGE_NAME}:${IMAGE_TAG}"
    
    log "To run the container:"
    echo "  limactl shell $LIMA_INSTANCE_NAME -- nerdctl run -p 8080:8080 ${IMAGE_NAME}:${IMAGE_TAG}"
    
    log "Note: The Lima instance will be automatically cleaned up when this script exits"
}

# Run main function
main "$@"