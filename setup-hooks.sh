#!/usr/bin/env bash
#
# Setup Git Hooks
# Run this after cloning the repository to enable pre-commit validation
#

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "Setting up Git hooks for cortex-wiki..."
echo

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}❌ Error: Not in a git repository${NC}"
    echo "Please run this script from the root of the cortex repository"
    exit 1
fi

# Check if .githooks directory exists
if [ ! -d ".githooks" ]; then
    echo -e "${RED}❌ Error: .githooks directory not found${NC}"
    echo "This script should be run from the repository root"
    exit 1
fi

# Make hooks executable
chmod +x .githooks/*

# Configure git to use .githooks
git config core.hooksPath .githooks

# Verify configuration
HOOKS_PATH=$(git config --get core.hooksPath)

if [ "$HOOKS_PATH" = ".githooks" ]; then
    echo -e "${GREEN}✓${NC} Git hooks configured successfully!"
    echo
    echo "Commit message validation is now active and will:"
    echo "  • Validate semantic versioning format before each commit"
    echo "  • Check commit message structure (type, scope, subject)"
    echo "  • Ensure header is ≤ 100 characters"
    echo "  • Verify subject doesn't end with a period"
    echo "  • Provide helpful error messages for invalid commits"
    echo
    echo -e "${YELLOW}Alternative: npm-based hooks${NC}"
    echo "If you prefer using Husky (npm-based hooks), run:"
    echo "  npm install"
    echo
    echo -e "${GREEN}Setup complete!${NC}"
    echo
    echo "Try making a commit to see the hook in action:"
    echo "  git commit -m \"test(hooks): verify commit validation\""
else
    echo -e "${YELLOW}⚠${NC} Warning: Hooks may not be configured correctly"
    echo "Expected: .githooks"
    echo "Got: $HOOKS_PATH"
    exit 1
fi
