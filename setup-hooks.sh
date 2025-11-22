#!/usr/bin/env bash
#
# Setup Git Hooks using pre-commit framework
# Run this after cloning the repository
#

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Setting up Git hooks (pre-commit)${NC}"
echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo

# Check if we're in a git repository
if [[ ! -d ".git" ]]; then
    echo -e "${RED}❌ Error: Not in a git repository${NC}"
    exit 1
fi

# Check if pre-commit is installed
if ! command -v pre-commit >/dev/null 2>&1; then
    echo -e "${RED}❌ pre-commit is not installed${NC}"
    echo
    echo "Install with one of:"
    echo "  pip install pre-commit"
    echo "  brew install pre-commit"
    echo
    exit 1
fi

echo -e "${GREEN}✓${NC} pre-commit is installed"
echo

# Remove old .githooks configuration if it exists
if git config --get core.hooksPath >/dev/null 2>&1; then
    OLD_PATH=$(git config --get core.hooksPath)
    echo -e "${YELLOW}⚠${NC}  Removing old hooksPath configuration: ${OLD_PATH}"
    git config --unset core.hooksPath
fi

# Install pre-commit hooks
echo "Installing pre-commit hooks..."
pre-commit install --install-hooks
pre-commit install --hook-type commit-msg

echo
echo -e "${GREEN}════════════════════════════════════════════${NC}"
echo -e "${GREEN}  ✓ Git hooks installed successfully!${NC}"
echo -e "${GREEN}════════════════════════════════════════════${NC}"
echo
echo "Active hooks:"
echo "  • Conventional commit message validation"
echo "  • Code formatting (Prettier, ESLint)"
echo "  • Python linting (Ruff)"
echo "  • Markdown linting"
echo "  • Secret detection"
echo "  • File checks (trailing whitespace, etc.)"
echo
echo -e "${BLUE}Usage:${NC}"
echo "  Automatic:  Hooks run on 'git commit'"
echo "  Manual:     pre-commit run --all-files"
echo "  Update:     pre-commit autoupdate"
echo "  Skip:       git commit --no-verify (not recommended)"
echo
echo -e "${GREEN}Setup complete!${NC}"
