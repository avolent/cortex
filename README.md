# cortex

Tiny Go binary that turns any folder of markdown into a navigable wiki. Point it at a repo, get a sidebar of every `.md` file and rendered HTML at `localhost:8090`. Zero config.

## Install

```bash
go install github.com/avolent/cortex@latest
```

Or build from source:

```bash
git clone https://github.com/avolent/cortex
cd cortex
go build
```

## Use

```bash
cortex                              # serve current directory on 127.0.0.1:8090
cortex -dir ~/path/to/repo          # serve another folder
cortex -dir ~/repo -addr 127.0.0.1:9000  # custom port
cortex -dir ~/repo -addr :8090      # bind all interfaces (LAN-accessible)
cortex -dir ~/repo -index index,home  # fall back to index.md/home.md when a folder has no README

cortex -dir ~/notes -export ./out   # render to static HTML, exit
```

Open `http://localhost:8090`. Or upload the `-export` output to any
static host (GitHub Pages, Netlify, S3). The default bind is loopback
only; pass `-addr :8090` to expose on the LAN.

## How it works

- Walks the directory recursively, collecting every `.md` file. Skips `.git`, `node_modules`, `vendor`, `tmp`, `dist`, `build`, `.next`, `.cache`, `.claude`, and any dotfile/dotdir.
- `/` serves the root `README.md`.
- `/path/to/foo` serves `path/to/foo.md`.
- `/path/to/foo.md` also works, so relative `[link](other.md)` references in your markdown resolve naturally.
- `/path/to/dir` falls back to `path/to/dir/README.md` if present, else the first `-index` name that matches (see `-index` flag).
- Embedded images (`.png .jpg .jpeg .gif .svg .webp .ico`) are served as static assets.
- Sidebar shows folders as collapsible `<details>` blocks. README is pinned to the top of each folder; everything else is alphabetical.
- Light + dark theme via `prefers-color-scheme`.
- **Live reload:** edits to any `.md` or image trigger an automatic browser reload via Server-Sent Events. New folders are picked up automatically.
- **Static export** (`-export <dir>`): same renderer writes the whole wiki to disk as HTML for hosting on GitHub Pages, Netlify, S3, etc.

## Out of scope

- Search (use `Ctrl-F`).
- Syntax highlighting in code blocks.
- Frontmatter, custom themes, config files.

If you need any of that, use Hugo or MkDocs.

## Deps

- `github.com/gomarkdown/markdown` for rendering.
- `github.com/fsnotify/fsnotify` for live reload.
