# GitHub Repository Settings

This document outlines recommended GitHub repository settings for optimal automation and security.

## Required Settings for Automation

### 1. Enable Renovate Bot

**Option A: GitHub App (Recommended)**

1. Go to https://github.com/apps/renovate
2. Click "Install"
3. Select your repository
4. Renovate will automatically detect `renovate.json` and start creating PRs

**Option B: Self-hosted**

- Follow instructions at https://docs.renovatebot.com/getting-started/running/

### 2. Branch Protection Rules

Enable for `main` branch:

**Settings → Branches → Add branch protection rule**

```
Branch name pattern: main

☑ Require a pull request before merging
  ☑ Require approvals: 0 (for personal repo) or 1 (for team)
  ☑ Dismiss stale pull request approvals when new commits are pushed

☑ Require status checks to pass before merging
  ☑ Require branches to be up to date before merging
  Required status checks:
    - Lint & Type Check
    - Build
    - Python Linting

☑ Require conversation resolution before merging

☐ Require signed commits (optional - recommended for security)

☑ Require linear history (keeps git history clean)

☐ Do not allow bypassing the above settings (optional)
```

### 3. GitHub Pages Settings

**Settings → Pages**

```
Source: Deploy from a branch
Branch: gh-pages (or as configured in Actions)
Folder: / (root)

Custom domain: avolent.io
☑ Enforce HTTPS
```

### 4. Security Settings

**Settings → Security → Code security and analysis**

```
☑ Dependency graph (should be enabled by default)
☑ Dependabot alerts
☑ Dependabot security updates
☐ Dependabot version updates (we use Renovate instead)
```

**Settings → Secrets and variables → Actions**

Add any required secrets (currently none needed for this repo)

### 5. Actions Settings

**Settings → Actions → General**

```
Actions permissions:
○ Allow all actions and reusable workflows

Workflow permissions:
○ Read and write permissions
☑ Allow GitHub Actions to create and approve pull requests (for Renovate auto-merge)
```

### 6. General Settings

**Settings → General**

```
Features:
☑ Issues
☐ Projects (optional)
☑ Discussions (optional)
☑ Wiki (optional - we use our own site)

Pull Requests:
☑ Allow merge commits
☐ Allow squash merging
☐ Allow rebase merging
☑ Always suggest updating pull request branches
☑ Automatically delete head branches
```

## Automation Workflow

With these settings enabled:

1. **Renovate** creates PRs for dependency updates weekly
2. **CI workflow** runs on all PRs (lint, type-check, build)
3. **Auto-merge** happens for minor/patch updates if CI passes
4. **Manual review** required for major updates
5. **Deploy workflow** runs on merge to main
6. **Pre-commit hooks** catch issues before commit (if developer has them installed)

## Testing the Setup

After configuring:

```bash
# Create a test branch
git checkout -b test/ci-check

# Make a small change
echo "# Test" >> test.md

# Commit and push
git add test.md
git commit -m "test: verify CI workflow"
git push origin test/ci-check

# Create a PR on GitHub
# Verify CI checks run automatically
```

## Monitoring

**Regular checks:**

- [ ] Review Renovate PRs weekly
- [ ] Check GitHub Actions status
- [ ] Monitor security alerts
- [ ] Review failed CI runs

**Dashboard access:**

- Actions: https://github.com/avolent/cortex/actions
- Security: https://github.com/avolent/cortex/security
- Insights: https://github.com/avolent/cortex/pulse

## Troubleshooting

### Renovate not creating PRs

- Check if GitHub App is installed
- Verify `renovate.json` is valid (run `npx renovate-config-validator`)
- Check Renovate logs in GitHub Actions

### CI checks not running

- Verify `.github/workflows/ci.yml` exists
- Check Actions are enabled in repository settings
- Review workflow permissions

### Auto-merge not working

- Ensure "Allow GitHub Actions to create and approve pull requests" is enabled
- Check that required status checks are passing
- Verify branch protection rules are configured correctly
