<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/fx-logo-dark.svg" />
    <img src="assets/fx-logo.svg" alt="fx — Fragment eXchange" width="420" />
  </picture>
</p>

<p align="center">
  <strong>Your server already knows how to render the page.<br />
  fx swaps the parts that changed.</strong>
</p>

<p align="center">
  One attribute · ~400 lines · no build step · no dependencies
</p>

<p align="center">
  <a href="https://fx.ciuffolo.com">The paper</a> ·
  <a href="https://fx.ciuffolo.com/patterns">Patterns</a> ·
  <a href="https://fx.ciuffolo.com/reference">Reference</a> ·
  <a href="https://fx.ciuffolo.com/demo">Live demo</a>
</p>

---

```html
<script src="/fx.js"></script>

<a href="/reports" fx-target="#content">Reports</a>

<main id="content">
  <!-- whatever your server rendered -->
</main>
```

Click the link and fx fetches `/reports`, lifts `#content` out of the response, puts it in
the page and updates the address bar. No reload, no flash, no scroll jump.

Turn JavaScript off and the same link still works, because it is a link to a page your
server renders. That is not a fallback bolted on afterwards — it is the starting point.

> A site built with fx is a site that works without fx. The library only makes it quicker.

## Why

Most internal tools, dashboards and admin panels are server-shaped: every fact on the
screen lives in a database the server can already reach. They still get built with a
package manager, a bundler, a component framework, a client-side router and a second copy
of the domain model, and none of that buys anything except a nicer transition.

fx is the smallest thing that gets you the nicer transition without the rest. HTMX and
Unpoly are both good software that did not fit — the long version, including what fx
borrows from Unpoly and what this approach costs you, is the paper at
[fx.ciuffolo.com](https://fx.ciuffolo.com).

## What it does

| | |
|---|---|
| `fx-target` | On a link or form: fetch the destination, swap these selectors |
| `fx-loading-target` | Add the `fx-loading` class to these while the request is in flight |
| `fx-hungry` | On any element with an `id`: swap me on every navigation |
| `fx-scroll` | Override where the page lands: `top`, `preserve`, or a selector |
| `<meta name="fx-refresh" fx-interval>` | Re-fetch this page on an interval, swap a fragment |

Where the page lands is decided by the URL, and usually you never touch it. A path change is
a new page, so it starts at the top; a `#hash` goes to that element. A query string change is
the same page refined — sorted, filtered, paged — so it stays where it was. `fx-scroll` is
for the case the URL cannot tell apart: pager links whose results should start at the top,
`fx-scroll="#results"`.

Plus one thing on your side: every request carries an `FX-Target` header naming the
fragments the browser is about to use, so a handler can skip work nobody asked for. It is
optional — a handler that ignores it is correct, just slower.

That is the whole API. If it were longer, the library would have failed.

## Install

```bash
curl -O https://fx.ciuffolo.com/fx.js
```

Copy it into your project. There is no package, no lockfile and no transitive dependency to
audit, because there is nothing to depend on. That URL always serves the latest release, so
what you get today is what you keep — the file is yours once you have it.

If you would rather not vendor it, tagged releases are on jsDelivr:

```html
<script src="https://cdn.jsdelivr.net/gh/kilianc/fragment-exchange@v1.1.2/fx.js" integrity="sha384-1S/m8Z1xZvmzHlvgjXz+CQEOr5YPEQGOYZyd+VwTFjzS1Ap3L3RrIXTmT8rvbc/A" crossorigin="anonymous"></script>
```

Pin the tag and keep the hash. An unpinned URL hands someone else the right to change the
JavaScript on your pages whenever they deploy, and the `integrity` attribute is what makes
the pin mean anything — the browser refuses the file if a single byte differs. Tags are
immutable, so both stay true forever. Vendoring is still the better answer if you have
somewhere to put the file: it is one request you own rather than one you rent.

`fx.dev.js` is an optional development companion: it turns on logging and warns about the
mistakes that are otherwise silent — duplicate ids, a hungry element with no id, a target
that matches nothing. Load it in development, never in production.

## Versioning

Semver, and the version covers three things:

| | |
|---|---|
| The attributes | `fx-target`, `fx-loading-target`, `fx-hungry`, `fx-scroll`, `fx-interval` |
| The `fx` object | `fx.timeout`, `fx.clickFallback`, `fx.version` |
| The `FX-Target` header | What the browser sends, and what a handler may assume |

That third one is easy to forget and the most expensive to get wrong. A server reading
`FX-Target` cannot be redeployed in lockstep with a script already sitting in someone's
browser cache, so changing what that header means is a breaking change even when the
markup and the runtime object are untouched.

Being in `1.x` is the promise that follows from that: none of the three breaks without a
`2.0.0`. Attributes may be added — `fx-scroll` arrived in `1.1.0` — but markup written
against an earlier `1.x` keeps working.

## Go

Optional. fx works with any server; this saves Go users writing the same twenty lines.

```bash
go get github.com/kilianc/fragment-exchange
```

```go
import fx "github.com/kilianc/fragment-exchange"

mux.Handle("GET /fx.js", fx.Handler())      // served from your binary

func reports(w http.ResponseWriter, r *http.Request) {
	page := ReportsPage{Filters: parseFilters(r)}

	// True on any page load, and on a fragment request that named it.
	if fx.Wants(r, "#results") {
		page.Results = search(r)     // skipped when only the nav is being swapped
	}

	render(w, page)
}
```

## Tests

```bash
go test ./...
```

One command, standard tooling, no Node and no test framework — nothing in `go.mod` exists
for the tests. Three suites run:

| | |
|---|---|
| `protocol_test.go` | The Go package: header parsing, `Wants`, the file handler |
| `fx_test.go` + `fx.test.js` | The library, in a real browser |
| `web/handler_test.go` | The site end to end, including that `FX-Target` really does skip work |

The browser suite is a web page, because fx is a few hundred lines of DOM, history and
`fetch` and there is nothing worth testing outside a browser. `fx_test.go` serves the
repository, opens `fx.test.html` in headless Chrome and waits for it to beacon its results
back — then reports **every case as a Go subtest**, so `-v`, `-run`, `-race`, `-json` and
CI annotations all work the way they do anywhere else.

```bash
make test-one FILTER=popstate   # one case
make test-headed                # watch it in a visible browser
make serve-tests                # drive the page yourself, with devtools open
```

`FX_CHROME` picks a browser. `FX_REQUIRE_CHROME=1` turns a missing browser into a failure
instead of a skip — CI sets it, so the suite can never quietly test nothing. A stray
`only:` in `fx.test.js` fails CI for the same reason.

## Repository

```text
fx.js            the library
fx.dev.js        development companion
fx.go            Go package: embeds the script, answers the FX-Target header
fx.test.js       the suite       fx.test.html   the page that runs it
fx_test.go       drives Chrome   protocol_test.go
web/             fx.ciuffolo.com, written in GSX and powered by fx
cmd/server/      the site's binary, same one locally and in production
```

```bash
make dev      # run the site on :8080
make test     # everything
make check    # what CI runs
```

The site is written in [GSX](https://github.com/kilianc/gsx) and served by a Go binary, so
`.gsx` files are source and `.gsx.go` files are generated — run `make gen` after editing
one.

## Releasing

A release is a git tag. There is no registry, no publish step and no build artefact — the
tag is the release, jsDelivr serves it from the tag, and `fx.ciuffolo.com/fx.js` follows
whatever is on `main`.

1. Bump `version` in `fx.js`.
2. Update the pinned URLs and `integrity` hashes wherever they appear — `README.md` and the
   site. `make sri` prints the current ones.
3. Write the `CHANGELOG.md` entry.
4. `bin/check-release v1.0.0`, then commit and merge to `main`.
5. `git tag v1.0.0 && git push origin v1.0.0`.

`bin/check-release` runs in CI on every pull request, and again on the tag push with the
tag name passed in. It fails if the declared version, the changelog, the published hashes
or the pinned URLs disagree with each other or with the file being shipped. That check
matters most on the tag: jsDelivr serves tags immutably, so a wrong hash published under
`v1.0.0` cannot be fixed in place — it costs a version number.

## Deploying the site

[fx.ciuffolo.com](https://fx.ciuffolo.com) is `cmd/server` on Vercel's Go preset: an
ordinary `net/http` server listening on `$PORT`. Same binary locally and in production —
no serverless shim, nothing that only exists in one environment.

`vercel.json` is the entire configuration:

```json
{ "framework": "go" }
```

Generated `.gsx.go` files are committed, so the build is a plain `go build` needing no
toolchain beyond Go. The project is connected to this repository, so pushing to `main`
deploys.

For anything the dashboard cannot do, the Vercel CLI runs from a container:

```bash
bin/vercel deploy --prod
```

There is no Node on this machine and there is not going to be. `tools/Dockerfile` pins the
toolchain — Node 22.12.0, Vercel CLI 58.9.4 — and `bin/vercel` mounts exactly two things
into it: this repository, and `.vercel-auth/` for the CLI's own credentials. Not `$HOME`,
not `~/.ssh`, not `~/.aws`. `make tools-image` rebuilds it; the wrapper does it for you on
first use.

## License

MIT © Kilian Ciuffolo
