# Agent Guidelines

## What this repository is

`fx.js` is the product. Everything else — the Go package, the tests, the site — exists to
serve it. The library's only real feature is that it is small enough to read in one
sitting, so **every change that grows it has to earn its lines**.

Before adding anything to `fx.js`, answer: can the user do this in their own five lines of
JavaScript, or in their server template? If yes, they should.

## File types

- `.gsx` files are **source** — edit these.
- `.gsx.go` files are **generated** by [GSX](https://github.com/kilianc/gsx) — never edit
  them directly. Run `make gen` after changing a `.gsx` file and commit both.
- `fx.js` and `fx.dev.js` are hand-written, plain, dependency-free browser JavaScript. No
  build step, no transpiler, no module system. They must stay copy-pasteable.

## Design rules, in priority order

1. **Simplicity beats performance.** fx re-fetches whole pages and discards most of the
   response. That is the deal. If an optimisation costs clarity, it belongs in the user's
   handler via the `FX-Target` header, not in the library.
2. **The server owns the page.** fx never composes, templates, or decides what a page
   contains. It reads a document and moves elements out of it.
3. **The URL is the state.** Anything that makes state unreachable by URL is a bug.
4. **Fail into the browser.** Every failure path ends with an ordinary navigation.
5. **No dependencies, ever.** Not in `fx.js`, not in `go.mod` for the tests. `go.mod`
   carries exactly one entry, for the site's rendering.

## Tests

The suite is `fx.test.js`, running in a real browser. `fx_test.go` serves the repository,
opens `fx.test.html` in headless Chrome and reports each case as a Go subtest.

- Any change to `fx.js` needs a test in `fx.test.js` first, red before green.
- Tests drive real events (`.click()`, `.requestSubmit()`) and wait for the DOM. Do not
  reach into fx's internals — if a behaviour cannot be observed from the outside, it is not
  a behaviour worth having.
- `make test-headed` opens a visible browser. `make serve-tests` lets you drive it
  yourself.
- Never leave an `only:` prefix in a committed test; CI fails on it, deliberately.

## Commit messages

```
scope: lowercase imperative description
```

Scopes: `fx`, `dev`, `go`, `site`, `test`, `ci`, `docs`. No trailing period, title under 72
characters, body after a blank line if it needs one.

Examples:

- `fx: keep multipart encoding when a form has a file`
- `site: name Unpoly as the source of fx-hungry`
- `test: cover the submitter's formaction and formmethod`

## The site

`web/` is `fx.ciuffolo.com`. It is server-rendered Go that uses fx, which means it is also
the largest integration test — if a change breaks real usage, the demo stops working.

Keep it honest: the demo's server log shows exactly what each request asked for and what
the handler chose to render. Do not fake a number in it.
