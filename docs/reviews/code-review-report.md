# cortex — Code Review Report

**Reviewer:** Application-stack best-practices audit
**Date:** 2026-07-31
**Commit state:** working copy at `/home/orc/git/cortex` (review-first; no changes committed)
**Scope:** `main.go`, `main_test.go`, `go.mod`/`go.sum`, `doc/`, `README.md`,
`.github/workflows/pages.yml`, `renovate.json`

## How this review was verified

All claims below were checked against the source directly. Where a fact is
cited (function name, constant, flag, line), it was grepped and confirmed at
`file:line`. Local checks were run read-only:

- `gofmt -l .` — clean (no unformatted files)
- `go vet ./...` — clean (no diagnostics)
- `go test ./...` — pass; `go test -cover` reports **63.7%** of statements
- `go mod tidy` (against a scratch copy, restored afterward) — would remove two
  stale `go.sum` lines (see Dependencies)

The server binary was **not** run, per the working-copy rules.

---

## Overall assessment

cortex is a genuinely well-executed small tool. It holds firmly to its own
stated constraints (one file, one binary, no config, minimal deps), the code
is readable end-to-end, error handling is consistent, and — importantly — every
security claim in `doc/architecture.md` still matches the implementation. There
is **no doc drift on the safety surface**, which is the headline result of this
review.

The gaps are the ones you would expect in a tool that has deliberately traded
breadth for simplicity: no graceful shutdown, no test coverage on the
concurrency/live-reload surface (including the SSE cap it advertises as a safety
feature), a near-empty Renovate config, and a couple of minor module-hygiene and
CI polish items. None are severe; most are low priority by design intent.

**Overall grade: strong.** The project does the hard things (safety, focus,
readability) well and leaves only low-risk polish on the table.

---

## 1. Project structure & Go conventions

**Strengths**

- The "one file" rule (`doc/rules.md` §1) is genuinely upheld: all Go code is in
  `main.go` (899 lines) and `main_test.go`. No `internal/`, no sub-packages, no
  separate template or asset files — the HTML/CSS/JS lives in the inline
  `tmplStr` constant (`main.go:59`), consistent with the "one binary" rule.
- `go.mod` is clean and minimal: module path `github.com/avolent/cortex`, Go
  `1.26.2`, exactly the two runtime deps named in the docs plus one indirect
  (`golang.org/x/sys`). Matches `AGENTS.md` and `doc/rules.md` §5.
- `gofmt` and `go vet` are both clean.

**Recommendations**

- (Low) `doc/rules.md` §2 states runtime assets are "embedded via `//go:embed`".
  In practice nothing uses `//go:embed` — the template is a plain inline `const`
  string, which is arguably *better* (nothing to embed). Reword §2 to say assets
  are inlined/embedded so a reader grepping for `go:embed` isn't confused. Docs
  finding, not a code finding.

**Priority: low**

---

## 2. Code quality & maintainability

**Strengths**

- Error handling is consistent: request-path errors are logged with context and
  converted to `404` (`main.go:551`, `main.go:567`); export-path errors are
  wrapped with `%w` and propagated (`exportSite`), so `main()` can `log.Fatalf`
  on them. Nothing is silently swallowed that matters.
- The `fileIndex` cache (`main.go` `fileIndex`/`snapshot`/`invalidate`) is a
  clean concurrency design: `invalidate()` replaces the maps wholesale and
  readers hold a snapshot reference under `RLock`, so a reader never observes a
  map being mutated. This is the correct copy-on-write pattern.
- Functions are small, single-purpose, and well-commented (the doc comments on
  `safeResolve`, `rewriteMdLinks`, `stripFrontmatter` explain intent and edge
  cases). The code is readable in one sitting.

**Recommendations**

- (Low) `fileIndex.snapshot()` uses double-checked locking but does **not**
  re-check `x.ready` after taking the write lock. Two concurrent first requests
  can both run `collectMarkdown()` and both write the result. It is harmless
  (idempotent, last-writer-wins) but does duplicate a full tree walk under a
  cold cache. A re-check after `x.mu.Lock()` removes the wasted walk. Efficiency
  nit only.
- (Low) `handler` rejects any cleaned path containing `..` via
  `strings.Contains(cleaned, "..")` (`main.go:543`). This also rejects
  legitimate filenames that merely contain a literal `..` (e.g. `a..b.md`).
  Given the target audience (your own repo) this is an acceptable, safe-by-
  default tradeoff — noted only so it is a conscious choice. The stronger guard
  is `safeResolve`, which runs regardless.

**Priority: low**

---

## 3. Security

Each `doc/architecture.md` "Safety" claim was checked against the code. **All
five still match** — no drift.

| Claim (architecture.md) | Code | Verdict |
|---|---|---|
| Path traversal: `path.Clean` + reject `..` | `main.go:542–544` (`path.Clean("/"+…)`, then `strings.Contains(cleaned, "..")` → `404`) | Matches |
| Symlink containment: `safeResolve()` / `EvalSymlinks` inside root | `safeResolve` (`main.go:224`) via `isUnderRoot`; used in `collectMarkdown`, `readMarkdown` (`main.go:518`), static-asset branch, and export | Matches |
| 10 MiB render cap | `maxFileSize = 10 << 20` (`main.go:33`), enforced at `main.go:535` with a defensive `io.LimitReader(f, maxFileSize+1)` at `main.go:538` | Matches |
| SSE connection cap 256 | `maxSSEClients = 256` (`main.go:34`); `subscribe` returns `nil` at `len(r.subs) >= maxSSEClients` (`main.go:598`) → `503` in `sseHandler` | Matches |
| Loopback-only default bind | `addr` default `127.0.0.1:8090` (`main.go` flag decl) | Matches |

**Additional strengths**

- Defence in depth: even if the `fileIndex` were stale, `readMarkdown` re-runs
  `safeResolve` per request, so an escaping symlink is refused at read time as
  well as walk time. `TestSymlinkEscapesRootRejected` exercises exactly this.
- The root itself is `EvalSymlinks`-resolved once at startup (`main.go:868`) so
  descendant containment comparisons are consistent (also mirrored in the test
  helper for macOS `/var → /private/var`).
- Static-asset serving is gated by an extension allow-list (`staticExts`) *and*
  `isSkipped` *and* `safeResolve` before `http.ServeFile`.

**Recommendations**

- (Low, by design) The trust model is explicitly "your own markdown"; inline
  HTML/`<script>` is passed through un-sanitised (`html.CommonFlags`, no
  `SkipHTML`). This is documented in `doc/usage.md` "Trust model" and is a
  deliberate choice, not a defect. No action needed beyond keeping that note
  prominent.
- (Low) There is no per-request body/time limit on the SSE endpoint beyond the
  connection count; a client can hold a slot open. The 256-cap bounds the blast
  radius and this is a loopback dev tool, so this is acceptable. Noted for
  completeness.

**Priority: low** (the security posture is sound; nothing here is a live risk)

---

## 4. Testing

**Strengths**

- Good pure-function coverage: `urlForFile`, `outPathForURL`, `isSkipped`,
  `titleForFile`, `rewriteMdLinks` (extensive table), `stripFrontmatter`
  (YAML/TOML/CRLF/unterminated/mismatched-fence edge cases), and
  `resolveMarkdown`.
- The safety features are meaningfully tested, not just asserted:
  - Symlink escape rejected — `TestSymlinkEscapesRootRejected`
  - In-root symlink accepted — `TestSymlinkInsideRootAccepted`
  - 10 MiB cap — `TestReadMarkdownSizeLimit`
  - Path traversal (`/../etc/passwd`, `/a/../../../etc/passwd`) and dotdir
    hiding (`/.git/config`) — `TestHandler`
- `TestExportSite` verifies the full static-export contract: file layout,
  link rewriting (`href="page/"`, `href="sub/"`), canonical sidebar URLs, and
  crucially that the exported HTML contains **no** `EventSource` live-reload
  script.

**Recommendations**

- (Medium) The **SSE connection cap (256) is untested**. It is advertised as a
  safety feature in `doc/architecture.md`, so it deserves a regression test.
  `reloader.subscribe`/`unsubscribe`/`broadcast` are unit-testable without a
  server: fill to `maxSSEClients`, assert the 257th `subscribe()` returns `nil`,
  then `unsubscribe` one and assert a slot frees. This closes the biggest
  coverage gap on the claimed safety surface.
- (Low) No test exercises `watchTree`/live-reload/debounce. This is genuinely
  awkward to test deterministically (filesystem timing), so leaving it uncovered
  is a reasonable call — but a single "broadcast reaches a subscribed channel"
  test on `reloader` alone (no fsnotify) would cover the fan-out logic cheaply.
- (Low) `exportSite`'s symlink-escape branch for **static assets** (images that
  resolve outside root are skipped) is not tested, though the equivalent `.md`
  path is. A small fixture with an escaping `.png` symlink would close it.
- (Low) Coverage is 63.7%. The uncovered statements are concentrated in `main`,
  `sseHandler`, `watchTree`, and the `reloader`/`copyAsset` error paths. Per the
  `coverage-ratchet` practice in `AGENTS.md`, adding the SSE-cap test would lift
  this; consider recording a floor once it moves.

**Priority: medium** (driven by the untested SSE cap)

---

## 5. Documentation

**Strengths**

- Documentation is accurate against the code. The three CLI flags (`-dir`,
  `-addr`, `-export`) in `doc/usage.md` and `README.md` exactly match the flag
  declarations in `main.go`. Defaults (`.`, `127.0.0.1:8090`, unset) match.
- The skipped-directory list in the docs (`.git`, `node_modules`, `vendor`,
  `tmp`, `dist`, `build`, `.next`, `.cache`, `.claude`) exactly matches
  `skipDirs` in `main.go`.
- URL-behaviour and export-layout tables match `urlForFile`/`outPathForURL`
  behaviour (and are backed by tests).
- `doc/rules.md` is a clear, enforceable statement of the project's constraints,
  and the code honours all of them.

**Recommendations**

- (Low) `doc/rules.md` §2 `//go:embed` wording — see §1 above (docs finding).
- (Low) The task brief and this report write to `docs/reviews/…`, but the
  project's own documentation folder is `doc/` (singular). This report
  intentionally follows the requested `docs/reviews/` path; if the project wants
  a single docs root, consider relocating this under `doc/reviews/` for
  consistency with the existing `doc/` convention.

**Priority: low**

---

## 6. CI/CD & DevOps

**Strengths**

- `.github/workflows/pages.yml` implements the correct gate order:
  `go test ./...` → `go build -o cortex .` → `./cortex -dir doc -export _site`
  → deploy. A failing test or build blocks the Pages deploy. This matches the
  `ci-test-build-gate` practice in `AGENTS.md`.
- Least-privilege permissions (`contents: read`, `pages: write`,
  `id-token: write`) and a `concurrency: { group: pages, cancel-in-progress:
  false }` guard so deploys don't race.
- `setup-go` uses `go-version-file: go.mod`, so CI can't drift from the declared
  toolchain. Actions are on current majors (`checkout@v6`, `setup-go@v6`,
  `deploy-pages@v5`).

**Recommendations**

- (Low) CI does not run `gofmt -l`/`go vet`. Both pass locally today, but a
  format/vet gate is cheap insurance against drift and aligns with the
  `run-local-checks` practice. Add a step, e.g.
  `test -z "$(gofmt -l .)"` and `go vet ./...`, before the build.
- (Low) Actions are pinned to major tags rather than commit SHAs. SHA-pinning is
  the hardening best practice for supply-chain integrity; optional for a
  low-risk docs deploy but worth noting.
- (Low) The `CNAME` is written inline (`echo cortex.wiki > _site/CNAME`) while a
  committed `CNAME` file also exists at the repo root. Not a bug (the export dir
  is separate), but the two sources of the same value could drift; consider
  copying the repo `CNAME` into `_site` instead of hard-coding the string.

**Priority: low**

---

## 7. Dependencies

**Strengths**

- The dependency set is exactly the documented ceiling: `fsnotify` and
  `gomarkdown` (runtime), `golang.org/x/sys` (indirect, pulled by fsnotify).
  Consistent with `doc/rules.md` §5 "minimal Go dependencies".
- `gomarkdown` has no tagged releases upstream, so the pseudo-version pin in
  `go.mod` is expected and correct, not sloppiness.

**Recommendations**

- (Low) **`go.sum` carries stale entries.** `go mod tidy` removes two lines for
  `github.com/fsnotify/fsnotify v1.9.0` (the module now requires `v1.10.1`).
  Running `go mod tidy` cleans this; it is purely hygiene (no security impact),
  but keeping `go.sum` tidy avoids confusion about which versions are in play.
- (Medium) **`renovate.json` is effectively empty** — it contains only
  `$schema` with no `extends`, no `packageRules`, no schedule:
  ```json
  { "$schema": "https://docs.renovatebot.com/renovate-schema.json" }
  ```
  Renovate will still auto-detect the Go module with built-in defaults, so it
  technically "tracks" deps, but there is no explicit configuration backing the
  `doc/rules.md` §5 intent. Recommend at minimum
  `"extends": ["config:recommended"]` so the behaviour is declared rather than
  implicit, and consider grouping/scheduling given the tiny, stable dep set. A
  short comment tying it to the "minimal deps" rule would make the intent
  self-documenting.
- (Low) `golang.org/x/sys v0.13.0` is somewhat behind current releases. It is an
  indirect dep of fsnotify and low-risk, but a configured Renovate would surface
  the bump.

**Priority: medium** (Renovate config) / low (go.sum, x/sys)

---

## 8. Industry standards

**Strengths**

- Sensible `http.Server` hardening for the workload: `ReadHeaderTimeout`
  (`main.go:893`) and `IdleTimeout` are set, and `WriteTimeout` is
  *intentionally* left unset with a comment explaining that SSE connections are
  long-lived (`main.go:895`). This shows the timeout choices were deliberate.
- Configuration is flags-only by design (`doc/rules.md` §3 forbids env vars and
  config files). So the "config from flags/environment" checklist item is a
  *deliberate non-goal*, not a gap — flags cover it and env is ruled out on
  purpose.

**Recommendations**

- (Medium) **No graceful shutdown / signal handling.** `main()` ends in
  `log.Fatal(srv.ListenAndServe())` (`main.go:898`), and `watchTree` is started
  with `context.Background()` (`main.go:884`) that is never cancelled. On
  `SIGINT`/`SIGTERM` the process dies immediately, dropping in-flight responses
  and open SSE connections without a clean close. For a local dev tool this is
  low-impact, but wiring `signal.NotifyContext` + `srv.Shutdown(ctx)` and
  passing that context into `watchTree` is the standard Go pattern and would
  make the already-present `ctx.Done()` branch in `watchTree` reachable. Small,
  self-contained, fits within `main.go`.
- (Low) **Logging is stdlib `log`, not structured.** Given Go 1.26, `log/slog`
  is in the stdlib (no new dependency), so structured logging is available
  without violating the deps rule. That said, `doc/rules.md` §7 explicitly
  favours simplicity, and the current log lines are perfectly readable — treat
  this as optional, not a real gap.
- (Low) **No `-version` flag / build-info.** A `-version` that prints
  `runtime/debug.ReadBuildInfo()` is a common nicety for a `go install`-
  distributed binary and costs nothing in dependencies. Optional.

**Priority: medium** (graceful shutdown) / low (logging, version flag)

---

## Priority-ranked action items

### Medium
1. **Add a regression test for the SSE connection cap (256).** It is an
   advertised safety feature and is currently untested. Unit-test
   `reloader.subscribe`/`unsubscribe` directly (§4). *(Testing)*
2. **Add graceful shutdown.** Use `signal.NotifyContext` + `srv.Shutdown` and
   pass the cancellable context into `watchTree` so its `ctx.Done()` path is
   actually used (`main.go:884`, `main.go:898`). *(Industry standards)*
3. **Configure Renovate explicitly.** Add `"extends": ["config:recommended"]`
   (and optionally grouping/schedule) to `renovate.json` so dependency tracking
   is declared, backing `doc/rules.md` §5. *(Dependencies)*

### Low
4. Run `go mod tidy` to drop the stale `fsnotify v1.9.0` lines from `go.sum`.
   *(Dependencies)*
5. Add a `gofmt`/`go vet` gate to `pages.yml` before the build step. *(CI/CD)*
6. Re-check `x.ready` after acquiring the write lock in `fileIndex.snapshot()`
   to avoid a duplicate cold-cache tree walk. *(Code quality)*
7. Test the static-asset symlink-escape branch in `exportSite`, and add a
   `reloader.broadcast` fan-out test. *(Testing)*
8. Fix `doc/rules.md` §2 `//go:embed` wording (template is an inline `const`,
   nothing is embedded). *(Documentation)*
9. Optional niceties: `-version` flag via `debug.ReadBuildInfo`; consider
   `log/slog`; copy the repo `CNAME` into `_site` rather than hard-coding the
   value in CI. *(Industry standards / CI)*

### None found
- No security drift: all five `doc/architecture.md` safety claims match the
  implementation.
- No formatting or `vet` issues; tests pass.

---

## Closing note

The most valuable finding here is a *negative* one: the safety documentation and
the code have not drifted apart, which is exactly the failure mode this review
was asked to catch. The remaining items are polish — the two worth doing soon
are the SSE-cap test (so a claimed safety feature is guarded against regression)
and graceful shutdown (the one genuine industry-standard gap). Everything else
is low-risk hygiene that can ride along with the next change, in keeping with the
project's `small-diffs` practice.
