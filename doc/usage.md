# Usage

## Install

```bash
go install cortex.wiki@latest
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
| `-addr` | `:8090` | Listen address (server mode only) |
| `-export` | (unset) | If set, render the wiki to static HTML in this directory and exit |

## Examples

```bash
cortex                          # serve current directory on :8090
cortex -dir ~/notes             # serve another folder
cortex -dir ~/repo -addr :9000  # custom port

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
| `/foo/` | `foo/README.md` if present |
| `/img/diagram.png` | static image, served as-is |

Lookups are case-insensitive on the file extension and stem, so `/TODO`
matches `TODO.MD`.

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

## Theme

Light + dark via the OS preference (`prefers-color-scheme`). No flag, no
toggle, no setting.
