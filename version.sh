#!/bin/bash

# Semantic version bumping script for Envoy
# Usage: ./version.sh [patch|minor|major]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print error and exit
error_exit() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

# Function to print success message
success_msg() {
    echo -e "${GREEN}$1${NC}"
}

# Check if a bump type is provided
if [ $# -eq 0 ]; then
    error_exit "Missing argument. Usage: $0 [patch|minor|major]"
fi

BUMP_TYPE=$1

# Validate bump type
case $BUMP_TYPE in
    patch|minor|major)
        ;;
    *)
        error_exit "Invalid bump type '$BUMP_TYPE'. Must be one of: patch, minor, major"
        ;;
esac

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    error_exit "Working directory is not clean. Commit or stash changes first."
fi

# Get the latest version tag
LATEST_TAG=$(git tag --sort=-version:refname --list "v[0-9]*" | head -n 1)

# If no tags exist, start with v0.0.0
if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG="v0.0.0"
    success_msg "No existing tags found. Starting with v0.0.0"
fi

# Remove 'v' prefix for processing
VERSION=${LATEST_TAG#v}

# Split version into components
MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)
PATCH=$(echo $VERSION | cut -d. -f3)

# Increment version based on type
case $BUMP_TYPE in
    patch)
        PATCH=$((PATCH + 1))
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
esac

# Construct new version
NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

# Confirm the bump
echo -e "${YELLOW}Bumping version:${NC} $LATEST_TAG -> $NEW_VERSION"
read -p "Continue? [y/N] " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Create annotated tag
git tag -a "$NEW_VERSION" -m "Bump version to $NEW_VERSION"

# Push the tag
git push origin "$NEW_VERSION"

success_msg "Successfully created and pushed tag: $NEW_VERSION"
echo -e "Users can now install with: ${GREEN}go install ytsruh.com/envoy@${NEW_VERSION}${NC}"
