# Rules

These are the hard rules for cortex. Breaking one needs an unusually strong
reason and an explicit conversation.

## 1. One file

All Go code lives in `main.go`. No packages, no sub-files, no `internal/`
directory. If a thing doesn't fit cleanly in `main.go`, it doesn't belong
in cortex.

The point: any contributor reads one file and has the whole program in
their head. No hunting across modules.

## 2. One binary

`go build` produces a single self-contained executable. No external
runtime files — no CSS dropped next to the binary, no fonts in a sibling
folder, no template files. Anything cortex needs at runtime is inlined
directly into `main.go` (the HTML/CSS/JS template is a plain `const`
string) and served from memory.

The point: `cortex` works after a single download. No setup, no asset
hunt.

## 3. No config files

Behaviour is controlled by CLI flags only. No `cortex.toml`, no
`.cortexrc`, no `$CORTEX_*` environment variables. Sensible defaults; a
flag for the few things that vary.

The point: zero-config means zero-config. Configuration is a feature
surface that grows; we don't open the door.

## 4. No external runtime dependencies

No CDN links in templates. No HTTP calls to other services from the
running binary. No outbound network, period.

Build-time / compile-time dependencies are fine: Go modules are vendored
and compiled in; static assets are embedded. The distinction is *runtime*:
once built, cortex works fully offline with zero outbound network.

The point: cortex is a tool you point at a folder. It should work on a
plane.

## 5. Minimal Go dependencies

The current dependency set (`fsnotify`, `gomarkdown`) is the ceiling.
Adding a dependency requires demonstrating it solves a real problem we
can't reasonably solve in `main.go` ourselves.

The point: every dependency is supply-chain risk, build-time cost, and
something to keep updated. We pay the cost only when we can't avoid it.

## 6. Stay focused

cortex's job: point at a folder, get a navigable markdown wiki with live
reload. That's it.

Out of scope:

- Search (browser `Ctrl-F` is fine)
- Plugins or extension points
- Custom themes via config
- Frontmatter as a feature (no titles, dates, or tags surfaced from it).
  A leading YAML (`---`) or TOML (`+++`) block is silently stripped
  before rendering so it doesn't show up as garbled output, but its
  contents are never read.
- User CSS injection
- Auth, multi-user, editing

In scope:

- Serving a folder as a markdown wiki with live reload (the core)
- Exporting that same wiki to static HTML (`-export`) so it can be
  hosted by GitHub Pages, Netlify, S3, or any other static host

If you need anything else, use Hugo or MkDocs. cortex is intentionally
narrower.

## 7. Performance and simplicity over features

Favour the obvious, efficient implementation over the flexible one. If
two approaches work, prefer the one that's easier to read.

We rebuild the nav tree on every request because it's fast enough and
caching is more complexity than it's worth. We accept this tradeoff
deliberately.
