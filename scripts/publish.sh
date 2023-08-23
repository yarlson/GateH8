#!/bin/bash

# Define variables
IMAGE_NAME="gateh8"
DOCKER_USERNAME="yarlson"
VERSION="0.1.0"

# Create a new builder instance
docker buildx create --name multiarchbuilder --use

# Bootstrap the builder with multi-architecture capabilities
docker buildx inspect multiarchbuilder --bootstrap

# Build and push the multi-architecture image
docker buildx build --platform linux/amd64,linux/arm64 -t "$DOCKER_USERNAME/$IMAGE_NAME:$VERSION" --push .

# Cleanup
docker buildx rm multiarchbuilder
