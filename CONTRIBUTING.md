# Contributing to Cortex Wiki

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Semantic Versioning & Commit Messages

This project follows [Semantic Versioning](https://semver.org/) and uses [Conventional Commits](https://www.conventionalcommits.org/) to ensure consistent versioning and changelog generation.

### Commit Message Format

All commits **must** follow this format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Examples:**

```
feat(sync): add dry-run mode to obsidian sync script

Adds --dry-run flag to preview what would be synced without making changes.
This helps users verify the sync before applying changes.

Closes #123
```

```
fix(layout): correct dark mode toggle accessibility

Added aria-label to dark mode button for screen readers.
```

```
docs(readme): update installation instructions

Added prerequisites section and improved setup steps.
```

### Commit Types

| Type       | Description                                             | Version Impact |
| ---------- | ------------------------------------------------------- | -------------- |
| `feat`     | A new feature                                           | Minor (0.X.0)  |
| `fix`      | A bug fix                                               | Patch (0.0.X)  |
| `docs`     | Documentation only changes                              | None           |
| `style`    | Code style changes (formatting, no logic change)        | None           |
| `refactor` | Code change that neither fixes a bug nor adds a feature | None           |
| `perf`     | Performance improvement                                 | Patch (0.0.X)  |
| `test`     | Adding or updating tests                                | None           |
| `build`    | Changes to build system or dependencies                 | None           |
| `ci`       | Changes to CI configuration                             | None           |
| `chore`    | Other changes that don't modify src or test files       | None           |
| `revert`   | Reverts a previous commit                               | Depends        |

### Breaking Changes

For breaking changes, add `BREAKING CHANGE:` in the footer or add `!` after type:

```
feat!: redesign obsidian sync API

BREAKING CHANGE: The sync script now requires Python 3.12+
and uses a different command-line interface.

Old: python bin/obsidian_sync.py /path/to/vault
New: python bin/obsidian_sync.py --obsidian-path /path/to/vault
```

This triggers a **major version bump** (X.0.0).

## Making a Commit

### Option 1: Interactive Commit (Recommended)

Use the interactive commit helper:

```bash
npm run commit
```

This will prompt you to fill in all required fields.

### Option 2: Manual Commit

Write the commit message manually:

```bash
git add .
git commit -m "feat(auth): add user authentication"
```

**Note:** Husky will validate your commit message automatically. If it doesn't follow the format, the commit will be rejected.

### Option 3: Traditional Commit (Will be Validated)

```bash
git add .
git commit
# Your editor will open
# Write your commit message following the format
# Commitlint will validate when you save and close
```

## Development Workflow

### 1. Fork and Clone

```bash
# Fork on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/cortex
cd cortex

# Install dependencies
npm install

# Setup git hooks (choose one method)
# Option 1: Standalone (no npm required)
./setup-hooks.sh

# Option 2: Husky (already installed via npm install)
# No additional action needed
```

**Note:** Git hooks are required to validate commit messages locally. See below for validation details.

### 2. Create a Branch

Use semantic branch names:

```bash
git checkout -b feat/add-search
git checkout -b fix/dark-mode-bug
git checkout -b docs/update-readme
```

### 3. Make Changes

```bash
# Make your changes
# Test locally
make dev

# Run quality checks
make lint
make type-check
make format
```

### 4. Commit with Semantic Messages

```bash
# Interactive (recommended)
npm run commit

# Or manual
git commit -m "feat(search): add full-text search functionality"
```

### 5. Push and Create PR

```bash
git push origin feat/add-search
```

Then create a Pull Request on GitHub.

## Pre-commit Hooks (Optional)

Install pre-commit hooks for automatic validation:

```bash
pip install pre-commit
pre-commit install
```

This will automatically:

- Validate commit messages
- Format code
- Run linters
- Check for common issues

## Pull Request Guidelines

### PR Title

Use the same format as commit messages:

```
feat(component): add new feature
fix(bug): resolve issue with X
docs: update contributing guide
```

### PR Description

Include:

- **What** changed
- **Why** it changed
- **How** to test
- **Screenshots** (if UI changes)
- **Breaking changes** (if any)
- **Issue references** (e.g., "Closes #123")

### Before Submitting

Ensure all checks pass:

```bash
# Run all quality checks
make lint && make type-check && make build

# Format code
make format

# Verify changes
git status
git diff
```

## Code Review Process

1. Create your PR
2. CI checks run automatically (lint, type-check, build, commit validation)
3. Reviewer provides feedback
4. Make requested changes
5. Once approved and CI passes, PR is merged

## Semantic Versioning Examples

### Version Number: X.Y.Z

```
Major.Minor.Patch
```

**Patch Release (0.0.X):**

```
fix(layout): correct spacing issue
perf(build): optimize image loading
```

**Minor Release (0.X.0):**

```
feat(search): add search functionality
feat(export): add PDF export
```

**Major Release (X.0.0):**

```
feat!: redesign entire UI

BREAKING CHANGE: Old theme configuration no longer works.
Users must update their config files.
```

## Commit Message Validation

### Valid Examples

✅ `feat(sync): add incremental sync mode`
✅ `fix: resolve dark mode flicker`
✅ `docs(readme): update installation steps`
✅ `chore(deps): update astro to v5`
✅ `refactor(utils): simplify date formatting`

### Invalid Examples

❌ `Added new feature` - Missing type and format
❌ `feat:add search` - Missing space after colon
❌ `FIX(bug): solve issue` - Type must be lowercase
❌ `feat(search): Added search.` - Subject shouldn't end with period
❌ `update readme` - Missing type

### Validation Rules

- ✅ Type must be from allowed list
- ✅ Type must be lowercase
- ✅ Subject must not be empty
- ✅ Subject must not end with period
- ✅ Header must be ≤ 100 characters
- ✅ Body lines must be ≤ 100 characters

## Questions?

- Check [README.md](./README.md) for general information
- Review [SECURITY.md](./SECURITY.md) for security practices
- See [.github/GITHUB_SETTINGS.md](./.github/GITHUB_SETTINGS.md) for CI/CD setup

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
