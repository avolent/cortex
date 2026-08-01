# Usage

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

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-dir` | `.` | Folder to serve (or export) |
| `-addr` | `127.0.0.1:8090` | Listen address (loopback by default; use `:8090` to expose on all interfaces) |
| `-export` | (unset) | If set, render the wiki to static HTML in this directory and exit |
| `-index` | (unset) | Comma-separated extra directory-index filenames to try when a directory has no README, e.g. `-index index,home` |
| `-version` | `false` | Print version info and exit |

## Examples

```bash
cortex                          # serve current directory on :8090
cortex -dir ~/notes             # serve another folder
cortex -dir ~/repo -addr :9000  # custom port
cortex -dir ~/repo -index index,home  # fall back to index.md/home.md when a folder has no README

cortex -dir doc -export ./out   # render docs to ./out, exit
```

Server mode: open `http://localhost:8090`.

## Static export

`-export <dir>` walks every `.md` file under `-dir`, renders it to HTML,
and writes the result under `<dir>` along with copies of any embedded
images. The output is a static site you can publish to GitHub Pages,
Netlify, S3, or anywhere that serves files.

Output layout:

| Source | Output |
|---|---|
| `README.md` | `out/index.html` |
| `foo.md` | `out/foo/index.html` |
| `foo/bar.md` | `out/foo/bar/index.html` |
| `foo/README.md` | `out/foo/index.html` |
| `img/diagram.png` | `out/img/diagram.png` |

URLs become directory-style (`/foo/`, `/foo/bar/`) so static hosts that
don't strip extensions still resolve them correctly.

Live reload is removed from the exported HTML — there's no server to
push events to.

## URL behaviour

| URL | Serves |
|---|---|
| `/` | `README.md` at the served root |
| `/foo` | `foo.md` |
| `/foo.md` | `foo.md` (so relative `[link](other.md)` works) |
| `/foo/` | `foo/README.md` if present, else the first matching `-index` name |
| `/img/diagram.png` | static image, served as-is |

Lookups are case-insensitive on the file extension and stem, so `/TODO`
matches `TODO.MD`.

If a directory has no `README.md`, pass `-index <names>` (comma-separated,
tried in order) to designate fallback index filenames — e.g. `-index
index,home` serves `foo/index.md` at `/foo` when `foo/README.md` doesn't
exist. README always wins if both are present. A configured `-index` name
collapses into its parent directory URL the same way `README.md` does, in
the sidebar and in `-export` output.

## Skipped paths

These directories are excluded from the sidebar and from URL routing:

`.git`, `node_modules`, `vendor`, `tmp`, `dist`, `build`, `.next`,
`.cache`, `.claude`

Any file or directory starting with a dot is also excluded.

## Live reload

Edits to any `.md` file or embedded image trigger an automatic browser
reload. New folders created while cortex is running are picked up
automatically — no restart needed.

## Sidebar layout

- Folders are collapsible `<details>` blocks (open by default)
- `README.md` is pinned to the top of each folder
- Everything else is alphabetical
- Collapsed/expanded state per folder is remembered in the browser's
  `localStorage`, so closing a folder keeps it closed across page loads,
  live-reloads, and navigation (works in both server mode and `-export`
  output)

## Theme

Light + dark via the OS preference (`prefers-color-scheme`). No flag, no
toggle, no setting.

## Trust model

cortex serves your own markdown unmodified. Like most CommonMark
renderers, gomarkdown passes through inline HTML — including `<script>`
tags — into the rendered page. That's fine for files you wrote yourself;
it means you should not point cortex at a folder of *untrusted*
markdown (e.g. user-submitted content) without sandboxing.

Symlinks are followed only if their targets stay inside the served
root. Symlinks pointing outside are silently dropped from both the
sidebar and URL routing.

The default listen address is `127.0.0.1:8090` so the wiki isn't
exposed on the local network. Pass `-addr :8090` (or `-addr 0.0.0.0:8090`)
to bind on all interfaces if you want LAN access.
