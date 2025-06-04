#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 [watch|build]"
    echo "  watch - Run Tailwind CSS in watch mode"
    echo "  build - Build Tailwind CSS with minification"
    exit 1
fi

case "$1" in
    "watch")
        echo "Starting Tailwind CSS in watch mode..."
        ./tailwindcss -i app.css -o ./static/css/index.css --watch
        ;;
    "build")
        echo "Building Tailwind CSS with minification..."
        ./tailwindcss -i app.css -o ./static/css/index.css --minify
        ;;
    *)
        echo "Error: Invalid parameter '$1'"
        echo "Usage: $0 [watch|build]"
        echo "  watch - Run Tailwind CSS in watch mode"
        echo "  build - Build Tailwind CSS with minification"
        exit 1
        ;;
esac
