---
cura:
  template: agents-md-baseline
  version: 1.4.0
  source: /home/orc/git/cura/templates/AGENTS.md.template
  last-verified: 2026-07-29
---

# AGENTS.md — cortex

This file gives AI agents the context and rules they need to work in this
project. It follows the [AGENTS.md open standard](https://agents.md) with
project conventions on top (cura-managed sections + Common/Project Practices
registries — see Practice types below).

The `base` section below is managed by [cura](https://github.com/avolent/cura).
Run `cura sync` to pull upstream changes. Everything outside the markers is
yours to edit freely.

How tools see this file in your project:

| Tool | File location |
|------|---------------|
| Cursor / Codex / Windsurf / Amp / Gemini CLI | `AGENTS.md` (natively read) |
| Claude Code | Symlink or copy to `CLAUDE.md` |
| GitHub Copilot | Symlink or copy to `.github/copilot-instructions.md` |

---

## Bootstrap — delete after filling in

Inspect the project (README, `go.mod`/`package.json`/etc., `.github/workflows/`)
to fill in any `{placeholders}` or `YYYY-MM-DD` values in this file —
**frontmatter included**. Run `cura lint` to verify, then delete this section.

---

## Project overview

cortex is a tiny zero-config Go binary that turns any folder of markdown
into a navigable wiki: point it at a repo and it serves a sidebar of every
`.md` file plus rendered HTML at `http://localhost:8090` (loopback by
default; `-addr :8090` exposes it on the LAN). It also renders the whole
tree to static HTML for hosting (`-export`), and live-reloads the browser
on file changes via SSE. Distributed via `go install`; its own docs site is
published to cortex.wiki through GitHub Pages.

## Tech stack

- **Language / runtime:** Go 1.26.2 (`go.mod`)
- **Build tool:** `go build` (Go toolchain; no Makefile)
- **Test framework:** Go standard `testing` (`main_test.go`)
- **CI/CD:** GitHub Actions (`.github/workflows/pages.yml`)
- **Hosting:** GitHub Pages for the docs site (cortex.wiki); the binary itself is distributed via `go install`
- **Key external services:** none — runtime deps are `github.com/gomarkdown/markdown` (rendering) and `github.com/fsnotify/fsnotify` (live reload)

## Build, test, and dev commands

```bash
go install github.com/avolent/cortex@latest  # install the released binary
go build -o cortex .                          # build from source
go test ./...                                 # run tests
gofmt -w .                                    # format
cortex -dir <path> -export ./out              # render the wiki to static HTML
```

---

## Practice types

Both the Common Practices and Project Practices tables below use the same
row Types:

- **Rule** — agent MUST enforce on every relevant action.
- **Convention** — preferred pattern; agent SHOULD follow.
- **Workflow** — there's a script or `make` target for this; follow Detail.
- **Validator** — automated check; agent shouldn't duplicate manually.
- **Reference** — informational; agent loads from Detail on demand.

Detail column points to authoritative source (file, anchor, script, or
URL). Use `—` if Summary fully captures the row.

**Both tables have a Category column** — a short lowercase label grouping
rows by the area of work they govern (e.g. `security`, `git`, `workflow`).
Use it to pull all relevant practices before acting in a given context
without having to read the whole table.

**How to use these tables:**

- **Before acting** (committing, pushing, modifying code, etc.), check
  both tables for relevant Rules and Conventions. Names are kebab-case
  and intent-revealing.
- **Look up a specific Practice by name:** `grep "<name>" AGENTS.md`
  (e.g. `grep "no-secrets" AGENTS.md`).
- **Filter by Type:** `grep "^| Rule" AGENTS.md` returns all Rules; same
  pattern for Convention / Workflow / Validator / Reference.
- **Filter by Category:** `grep "| security |" AGENTS.md` returns all
  security practices across both tables; same pattern for any category label.
- **Common Practices apply to any project**; Project Practices apply
  only here. Both are equally binding for work IN this project.

---

<!-- cura:begin base -->

## Common Practices

*This section is managed by [cura](https://github.com/avolent/cura).
Run `cura sync` to update from upstream. Don't edit manually — changes
will be overwritten on next sync.*

| Type      | Category  | Name                  | Summary                                                                                       | Detail |
|-----------|-----------|-----------------------|-----------------------------------------------------------------------------------------------|--------|
| Rule      | security  | no-secrets            | No API keys, credentials, tokens in commits                                                   | — |
| Rule      | git       | no-force-push-to-main | Don't push to main with `--force`                                                             | — |
| Rule      | git       | no-no-verify          | Don't bypass git hooks with `--no-verify`                                                     | — |
| Rule      | safety    | local-host-only       | All operations stay local to this host; no external mutations (pushes, deploys, API writes) without explicit approval | — |
| Rule      | workflow  | run-local-checks      | Run the project's local checks (tests, lint, format) before declaring a change done — where they exist | — |
| Rule      | quality   | verify-not-fabricate  | Verify facts (function names, env vars, versions); never fabricate                            | — |
| Rule      | quality   | grep-before-mention   | When citing a function/file/version, grep first and cite `file:line`                          | — |
| Reference | reference | agents-md-spec        | Project agent context standard                                                                | [agents.md](https://agents.md) |
| Rule      | git       | no-coauthored-by      | Don't add `Co-Authored-By` trailers to commit messages                                        | — |
| Rule      | quality   | small-diffs           | Keep changes minimal and focused; no unrelated cleanup riding alongside a fix                 | — |
| Rule      | docs      | docs-stay-current     | Update matching docs in the same change as the code; never let docs drift                     | — |
| Rule      | quality   | audit-every-site      | Before any repo-wide fix, grep the whole codebase for the pattern, list every affected site with `file:line`, confirm the list is complete, then fix all sites in the same commit | — |
| Rule      | testing   | test-error-messages   | Tests must assert the exact user-facing string (error text, label, content), not just a status code or a generic pass/fail | — |
| Rule      | quality   | evidence-before-conclusion | Never state a root cause or behaviour change as confirmed without real evidence (logs, source, diff); distinguish "confirmed" from "inferred" | — |
| Rule      | docs      | persist-research      | Any significant research, design, or decision reached in a conversation must be written to a durable doc before the session ends; conversation context is ephemeral | — |
| Rule      | testing   | tdd-confirm-red       | A red test must fail for the right reason (an assertion mismatch), not a typo/panic/nil-deref/setup bug; confirm the failure message before writing implementation | — |
| Rule      | testing   | test-first            | Write the test before the implementation; no feature or bug-fix code lands without a failing test first | — |
| Rule      | testing   | regression-on-fix     | Every bug fix ships with a test that reproduces the bug (fails first) before the fix is applied | — |
| Rule      | testing   | coverage-ratchet      | After a change raises test coverage, ratchet the coverage-threshold floor upward; only lower it with a logged justification | — |

<!-- cura:end base -->

---

## Project Practices

*Local additions specific to this project. Append rows freely. This
section is never touched by cura.*

Category is a short lowercase label for the area of work the row governs.
Suggested values: `git`, `testing`, `e2e`, `handler`, `security`,
`domain`, `docs`, `reference` — or any label that fits your project.

| Type       | Category  | Name                  | Summary                                | Detail |
|------------|-----------|-----------------------|----------------------------------------|--------|
| Rule       | {category} | example-project-rule | One-sentence rule statement            | — |
| Convention | {category} | example-convention   | Preferred pattern description          | [link] |
| Workflow   | {category} | example-workflow     | What it does and how to invoke         | [scripts/example.sh](scripts/example.sh) |
| Validator  | {category} | example-check        | What it checks                         | [scripts/validator.sh](scripts/validator.sh) |
| Reference  | reference  | example-reference    | Load when {trigger condition}          | [link-or-url] |
| Validator  | ci         | ci-test-build-gate   | CI (`pages.yml`) runs `go test ./...` then `go build -o cortex .` before rendering `doc/` to static HTML and deploying it to GitHub Pages (cortex.wiki) on push to main | [.github/workflows/pages.yml](.github/workflows/pages.yml) (lines 21-27) |
| Convention | release   | semver-release-tag   | After a user-facing change lands on `main`, tag a semver release and push it (`git tag vX.Y.Z && git push origin vX.Y.Z`): patch = fixes, minor = additive/backward-compatible features (e.g. a new flag), major = breaking CLI changes. Current latest: `v1.0.1`. This is what the `-version` flag's build-info output reflects — untagged commits print a pseudo-version. | — |
