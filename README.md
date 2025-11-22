<h1 align="center">Cortex Wiki</h1>

<p align="center">
  <em>A personal knowledge base built with Astro, synced from Obsidian</em>
</p>

<p align="center">
  <a href="https://avolent.io">View Live Site</a>
</p>

## About

The goal of this wiki is to be an accumulation and amalgamation of information from various topics. Using the Feynman Technique - building understanding through clear explanation using concise thoughts and simple language.

This is a collection of:

- Technical guides and tutorials
- Brain dumps from personal projects
- Curated notes and resources
- Development best practices

**Built with:**

- [Astro](https://astro.build/) - Static Site Generator
- [Latex.css](https://latex.vercel.app/) - Beautiful LaTeX-inspired styling
- [Obsidian](https://obsidian.md/) - Content management

## Features

- **Fast**: Built with Astro for optimal performance
- **SEO Optimized**: Auto-generated sitemap, semantic HTML
- **Dark Mode**: Persistent theme preference
- **Type-Safe**: TypeScript configuration
- **Quality Tooling**: ESLint, Prettier, and automated formatting
- **CI/CD**: Automated deployment via GitHub Actions
- **Responsive**: Mobile-friendly design
- **Accessible**: ARIA labels and semantic markup

## Quick Start

### Prerequisites

- Node.js 20+ (specified in `.nvmrc`)
- Python 3.12+ (for Obsidian sync)

### Installation

```bash
# Clone the repository
git clone https://github.com/avolent/cortex
cd cortex

# Install dependencies
make install

# Setup git hooks (pre-commit framework)
./setup-hooks.sh
# or
make hooks

# Start development server
make dev
```

The site will be available at `http://localhost:4321`

### Git Hooks Setup

This repository uses [pre-commit](https://pre-commit.com/) framework to enforce code quality and commit standards:

```bash
# Setup hooks (requires pre-commit to be installed)
./setup-hooks.sh
# or
make hooks
```

**What it does:**

- Validates commit messages (Conventional Commits)
- Formats code (Prettier, ESLint)
- Lints Python code (Ruff)
- Checks Markdown files
- Detects secrets and sensitive data
- Runs file checks (trailing whitespace, etc.)

**Installing pre-commit:**

```bash
pip install pre-commit
# or
brew install pre-commit
```

## Available Commands

```bash
make help          # Show all available commands
make install       # Install dependencies
make dev           # Start development server
make build         # Build for production
make preview       # Preview production build
make sync          # Sync content from Obsidian
make sync-dry-run  # Test sync without making changes
make lint          # Run ESLint
make format        # Format code with Prettier
make type-check    # Run TypeScript checks
make clean         # Remove build artifacts
make upgrade       # Upgrade Astro to latest
```

## Development Workflow

### 1. Local Development

```bash
# Start dev server with hot reload
npm run dev

# Or using make
make dev
```

### 2. Code Quality

```bash
# Lint code
npm run lint

# Auto-fix linting issues
npm run lint:fix

# Format code
npm run format

# Check types
npm run type-check
```

### 3. Build & Deploy

```bash
# Build for production
npm run build

# Preview production build locally
npm run preview

# Deploy (automatic via GitHub Actions on push to main)
git push origin main
```

## Obsidian Sync

Content is managed in Obsidian and synced to this repository using a custom Python script.

### Setup

```bash
# Set your Obsidian vault path
export OBSIDIAN_PATH="/path/to/your/vault"

# Or create a .env file (not committed)
echo 'OBSIDIAN_PATH="/path/to/vault"' > .env
```

### Usage

```bash
# Sync content from Obsidian
python3 bin/obsidian_sync.py

# Or use make
make sync

# Dry run to preview changes
make sync-dry-run

# Custom path
python3 bin/obsidian_sync.py --obsidian-path /custom/path

# See all options
python3 bin/obsidian_sync.py --help
```

The sync script:

- Copies markdown files and images from Obsidian
- Filters content based on `publish: "true"` frontmatter
- Updates internal links for web compatibility
- Handles images and cross-references

## Project Structure

```text
cortex/
├── bin/
│   └── obsidian_sync.py    # Obsidian sync script
├── public/
│   ├── dark-mode.js        # Dark mode toggle
│   ├── toc.js              # Table of contents
│   ├── style.css           # Custom styles
│   └── images/             # Synced images
├── src/
│   ├── layouts/
│   │   └── BaseLayout.astro
│   └── pages/              # Markdown content
│       ├── index.md
│       ├── devices/
│       ├── homelab/
│       ├── programming/
│       └── tools/
├── .github/
│   └── workflows/
│       └── deploy.yml      # GitHub Actions CI/CD
├── astro.config.mjs        # Astro configuration
├── tsconfig.json           # TypeScript config
├── .eslintrc.json          # ESLint config
├── .prettierrc.json        # Prettier config
├── package.json
├── Makefile
└── README.md
```

## Configuration

### Astro Config (`astro.config.mjs`)

- Site URL: `https://avolent.io`
- Integrations: MDX, Sitemap
- Build output: Static

### TypeScript (`tsconfig.json`)

- Strict mode enabled
- ES2022 target
- Full type checking

### Linting & Formatting

- **ESLint**: Code quality and best practices
- **Prettier**: Consistent code formatting
- Supports `.astro`, `.ts`, `.js` files

## Deployment

The site automatically deploys to GitHub Pages when changes are pushed to the `main` branch.

**GitHub Actions Workflow:**

1. Checkout code
2. Install dependencies
3. Build static site
4. Deploy to GitHub Pages

**Custom Domain:** Configured via `public/CNAME` file

## Automation & Maintenance

This repository uses several automated tools to maintain code quality and keep dependencies up-to-date.

### Renovate Bot

Automated dependency updates via Renovate:

- **Weekly updates** for all dependencies
- **Auto-merge** for minor/patch updates after CI passes
- **Manual review** required for major version updates
- **Security updates** prioritized and labeled

Configure Renovate:

1. Install the [Renovate GitHub App](https://github.com/apps/renovate)
2. Select this repository
3. Renovate will use `renovate.json` configuration

### Continuous Integration

All pull requests automatically run:

- ESLint code linting
- Prettier format checking
- TypeScript type checking
- Production build verification
- Python code linting (ruff)

See `.github/workflows/ci.yml` for configuration.

### Pre-commit Hooks (Optional)

Install pre-commit hooks for local development:

```bash
# Install pre-commit (requires Python)
pip install pre-commit

# Install the git hooks
pre-commit install

# Run manually on all files
pre-commit run --all-files
```

Pre-commit hooks will automatically:

- Format code with Prettier
- Lint JavaScript/TypeScript with ESLint
- Format and lint Python with Ruff
- Check for common issues (trailing whitespace, merge conflicts, etc.)
- Detect potential secrets

### GitHub Settings

See [.github/GITHUB_SETTINGS.md](./.github/GITHUB_SETTINGS.md) for recommended repository settings including:

- Branch protection rules
- Required status checks
- Auto-merge configuration
- Security settings

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/) and enforces [Conventional Commits](https://www.conventionalcommits.org/).

### Commit Message Format

All commits must follow this format:

```text
<type>(<scope>): <subject>
```

**Examples:**

```bash
feat(sync): add dry-run mode
fix(layout): correct dark mode toggle
docs(readme): update installation steps
```

### Making Commits

#### Option 1: Interactive (Recommended)

```bash
npm run commit
```

#### Option 2: Manual

```bash
git commit -m "feat(search): add full-text search"
```

**Validation:** Husky automatically validates commit messages. Invalid commits are rejected.

### Commit Types

- `feat` - New feature (minor version bump)
- `fix` - Bug fix (patch version bump)
- `docs` - Documentation changes
- `style` - Code formatting
- `refactor` - Code restructuring
- `perf` - Performance improvements
- `test` - Test updates
- `build` - Build system changes
- `ci` - CI configuration changes
- `chore` - Maintenance tasks

**Breaking changes:** Add `!` after type or `BREAKING CHANGE:` in footer (major version bump)

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

## Contributing

This is a personal wiki, but contributions are welcome!

**Quick Start:**

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes
4. Commit using semantic format (`npm run commit`)
5. Run quality checks: `make lint && make type-check && make build`
6. Push and create a Pull Request

**Important:** All commits must follow [Conventional Commits](https://www.conventionalcommits.org/) format. See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## Security

See [SECURITY.md](./SECURITY.md) for security considerations and best practices.

## License

Personal project - content and code by [avolent](https://github.com/avolent)

## Acknowledgments

- [Astro](https://astro.build/) - Amazing static site generator
- [Latex.css](https://latex.vercel.app/) - Beautiful typography
- [Obsidian](https://obsidian.md/) - Powerful note-taking
- The open-source community
