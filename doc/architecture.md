# Architecture

cortex is one Go file, one binary, no config. This document explains how the
pieces fit together.

## File layout

```
main.go     # entire program
go.mod      # module declaration + dependencies
README.md   # public-facing intro
doc/        # this folder
```

There are no packages, no internal subdirectories of Go code, no templates
or assets stored as separate files. Everything that runs at runtime lives
inside `main.go` or is embedded into the binary at build time.

## Request flow

1. **Process start** — `main()` parses flags (`-dir`, `-addr`), resolves the
   serve root to an absolute path, starts a goroutine watching the tree
   with `fsnotify`, and calls `http.ListenAndServe`.
2. **HTTP request arrives** at `handler()`:
   - The URL path is cleaned (removes `..`, normalises slashes).
   - `resolveMarkdown()` maps the URL to an on-disk `.md` file — trying
     `path.md`, `path/README.md`, and case-insensitive fallbacks.
   - If a markdown file is found, it's read, rendered to HTML via
     `gomarkdown`, wrapped in the page template, and written to the
     response.
   - If no markdown matches but the URL points to an embedded image
     (`.png`, `.jpg`, `.svg`, etc.) inside the served tree, `http.ServeFile`
     hands it back.
   - Otherwise: 404.
3. **Sidebar** is built from a cached file list (`fileIndex`). The list
   is collected once at first request and reused; the fsnotify goroutine
   invalidates it whenever a `.md` file is created, removed, or renamed.
   `buildTree` still recurses per request (it depends on which page is
   active) but no filesystem walk happens. README is pinned to the top
   of each folder; everything else alphabetical.

## Live reload

A second HTTP endpoint, `/_cortex/events`, holds open Server-Sent Event
connections. The fsnotify goroutine debounces filesystem events
(120 ms window) and broadcasts a `reload` message to every subscriber.
A one-line script in the page template listens and calls
`location.reload()`.

```
fsnotify event ─► debounce ─► broadcast ─► SSE ─► browser reload
```

New directories created after startup are watched automatically (the
event handler adds them on `Create`).

## Template

A single inline `tmplStr` constant holds the entire HTML page: layout,
CSS, the live-reload script, and a recursive sub-template for the nav.
There is no template file, no theme system, no plugin point.

## Static export mode

`-export <dir>` short-circuits the server entirely. `main()` calls
`exportSite()` and exits.

`exportSite()`:

1. Walks the tree (same `collectMarkdown()` used by the server) to find
   every `.md` file.
2. For each unique URL, calls `resolveMarkdown()` to apply the same
   precedence rules the HTTP handler uses (so `foo.md` beats
   `foo/README.md` for `/foo`).
3. Renders the page through the same template, with `Reload: false` so
   the SSE script is omitted (live reload requires a server).
4. Writes the HTML to `out/<path>/index.html` — directory-style URLs that
   work on hosts without extension stripping.
5. Walks the tree a second time copying any embedded static assets
   (images) to matching paths under `out/`.

The export reuses `collectMarkdown`, `resolveMarkdown`, `buildTree`,
`renderMarkdown`, and the HTML template — no parallel rendering path,
just a different sink for the output.

## Skipped paths

`isSkipped()` blocks any path segment that:

- Is in the hard-coded list (`.git`, `node_modules`, `vendor`, `tmp`,
  `dist`, `build`, `.next`, `.cache`, `.claude`)
- Starts with a dot

This keeps generated/heavy directories out of the sidebar and the URL
namespace.

## Safety

- **Path traversal**: incoming URLs are passed through `path.Clean` and
  rejected if they still contain `..`. Static-asset serving is gated by
  an extension whitelist.
- **Symlink containment**: `safeResolve()` calls `filepath.EvalSymlinks`
  and verifies the result stays inside `*rootDir`. A `.md` symlink that
  points to `/etc/passwd` is dropped at walk time and refused at request
  time.
- **Render size cap**: markdown files larger than 10 MiB are refused so
  a runaway file can't OOM the process.
- **SSE connection cap**: at most 256 concurrent live-reload subscribers.
  Excess connections get a `503` instead of being added to the
  broadcast list.
- **Default loopback bind**: `-addr` defaults to `127.0.0.1:8090` so the
  wiki isn't exposed on the LAN unless you explicitly opt in.

cortex assumes the markdown files it serves are trusted (your own
notes, your own repo). It does not sanitize inline HTML — see the
trust-model note in [usage.md](usage.md).

## Dependencies

- `github.com/gomarkdown/markdown` — CommonMark renderer
- `github.com/fsnotify/fsnotify` — cross-platform file watching

That's the entire ceiling. See [rules.md](rules.md) on adding more.
