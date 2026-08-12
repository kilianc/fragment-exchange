<h1 align="center">fx — Fragment eXchange</h1>

<p align="center">
  <strong>Your server already knows how to render the page.<br />
  fx swaps the parts that changed.</strong>
</p>

<p align="center">
  One attribute · ~400 lines · no build step · no dependencies
</p>

<p align="center">
  <a href="https://fx.ciuffolo.com">Documentation</a> ·
  <a href="https://fx.ciuffolo.com/why">Why it exists</a> ·
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
borrows from Unpoly and what this approach costs you, is at
[fx.ciuffolo.com/why](https://fx.ciuffolo.com/why).

## What it does

| | |
|---|---|
| `fx-target` | On a link or form: fetch the destination, swap these selectors |
| `fx-loading-target` | Add the `fx-loading` class to these while the request is in flight |
| `fx-hungry` | On any element with an `id`: swap me on every navigation |
| `<meta name="fx-refresh" fx-interval>` | Re-fetch this page on an interval, swap a fragment |

Plus one thing on your side: every request carries an `FX-Target` header naming the
fragments the browser is about to use, so a handler can skip work nobody asked for. It is
optional — a handler that ignores it is correct, just slower.

That is the whole API. If it were longer, the library would have failed.

## Install

```bash
curl -O https://fx.ciuffolo.com/fx.js
```

Copy it into your project. There is no package, no lockfile and no transitive dependency to
audit, because there is nothing to depend on.

`fx.dev.js` is an optional development companion: it turns on logging and warns about the
mistakes that are otherwise silent — duplicate ids, a hungry element with no id, a target
that matches nothing. Load it in development, never in production.

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

The suite is a web page, because fx is a few hundred lines of DOM, history and `fetch` and
there is nothing worth testing outside a browser. A Go test serves the repository, opens
the page in headless Chrome and waits for it to report back:

```bash
go test ./...
```

No Node, no test framework, nothing in `go.mod` for it. Set `FX_CHROME` to pick a browser,
`FX_HEADED=1` to watch it run, or just open `fx.test.html` yourself.

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

## License

MIT © Kilian Ciuffolo
