---
layout: /src/layouts/BaseLayout.astro
author: avolent
title: Python development setup
date: November 2024
publish: 'true'
---

## Summary

Modern Python development setup using UV - an extremely fast Python package installer and resolver written in Rust. UV replaces pip, pip-tools, pipenv, poetry, pyenv, and virtualenv with a single tool that's 10-100x faster.

## Assumptions

- Basic Python knowledge
- Works on macOS, Linux, and Windows

## Contents

1. [Why UV?](#why-uv)
2. [Installation](#installation)
3. [Basic Usage](#basic-usage)
4. [Project Setup](#project-setup)
5. [Python Version Management](#python-version-management)
6. [Useful Links](#useful-links)

## Why UV?

UV is a next-generation Python package manager that combines the functionality of multiple tools:

- **Fast**: 10-100x faster than pip
- **All-in-one**: Replaces pip, virtualenv, pyenv, poetry, and more
- **Simple**: Single binary, no Python required for installation
- **Modern**: Written in Rust, battle-tested dependency resolver

## Installation

### macOS and Linux

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### Windows

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### Using Homebrew (macOS/Linux)

```bash
brew install uv
```

### Verify Installation

```bash
uv --version
```

## Basic Usage

### Create a New Project

```bash
# Create a new project with virtual environment
uv init my-project
cd my-project

# Or initialize in existing directory
uv init
```

### Install Dependencies

```bash
# Install packages (automatically creates venv if needed)
uv add requests pandas numpy

# Install dev dependencies
uv add --dev pytest black ruff

# Install from requirements.txt
uv pip install -r requirements.txt
```

### Run Commands in Virtual Environment

```bash
# Activate virtual environment
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate     # Windows

# Or run commands directly without activation
uv run python script.py
uv run pytest
```

## Project Setup

### Initialize a New Project

```bash
# Create new project
uv init my-api
cd my-api

# Add dependencies
uv add fastapi uvicorn

# Add development tools
uv add --dev pytest ruff black mypy

# Create a simple app
cat > main.py << 'EOF'
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "World"}
EOF

# Run the app
uv run uvicorn main:app --reload
```

### Lock Dependencies

```bash
# Generate lock file for reproducible installs
uv lock

# Install from lock file
uv sync
```

## Python Version Management

UV can install and manage Python versions automatically:

```bash
# Install a specific Python version
uv python install 3.12

# List available Python versions
uv python list

# Use specific Python version for project
uv python pin 3.12

# This creates a .python-version file
# UV will automatically use this version
```

### Using Specific Python Version

```bash
# Create venv with specific Python version
uv venv --python 3.12

# Or specify when running
uv run --python 3.11 python script.py
```

## Common Workflows

### Development Workflow

```bash
# 1. Clone project
git clone https://github.com/username/project
cd project

# 2. Install dependencies (UV auto-creates venv)
uv sync

# 3. Run tests
uv run pytest

# 4. Format code
uv run black .
uv run ruff check --fix .

# 5. Add new dependency
uv add new-package
```

### Update Dependencies

```bash
# Update all packages
uv lock --upgrade

# Update specific package
uv lock --upgrade-package requests

# Sync to updated versions
uv sync
```

## Useful Commands

```bash
# Show installed packages
uv pip list

# Show dependency tree
uv tree

# Remove package
uv remove package-name

# Export requirements
uv pip freeze > requirements.txt

# Clean cache
uv cache clean
```

## Migration from Other Tools

### From pip + venv

```bash
# Before:
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# After:
uv venv
uv pip install -r requirements.txt
# or simply:
uv sync
```

### From Poetry

```bash
# UV can read pyproject.toml
uv sync
```

### From Pipenv

```bash
# Convert Pipfile to requirements
# Then use UV
uv pip install -r requirements.txt
```

## Useful Links

- [UV Documentation](https://docs.astral.sh/uv/)
- [UV GitHub](https://github.com/astral-sh/uv)
- [Astral - UV Creators](https://astral.sh/)
- [Ruff - Fast Python Linter](https://github.com/astral-sh/ruff)

## References

- [UV Official Docs](https://docs.astral.sh/uv/)
- [UV vs pip Benchmark](https://github.com/astral-sh/uv#benchmarks)
- [Python Packaging with UV](https://packaging.python.org/)
