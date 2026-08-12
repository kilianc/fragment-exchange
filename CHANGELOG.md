# Changelog

Releases are git tags. Each one is served from jsDelivr at
`cdn.jsdelivr.net/gh/kilianc/fragment-exchange@<tag>/fx.js`, immutably.

fx follows semver over three surfaces: the attributes, the `fx` object, and the
`FX-Target` header. See the versioning notes in the README for why the header
counts.

## 1.1.0

### The library

- `fx-scroll` on a link, a form or a submitter: `top`, `preserve`, or a selector
  to bring into view.
- Where a navigation lands is now decided by the URL. A path change still starts
  at the top and a `#hash` still goes to that element, but a query string change
  keeps the scroll position instead of resetting it: sorting, filtering and
  paging refine the page you are already reading rather than replacing it.

  Behaviour change for anyone on 1.0.0 whose query string navigations relied on
  the reset. `fx-scroll="top"` restores it for the links that want it.

## 1.0.0

First tagged release. The library, the Go package and the site as they stand.

### The library

- `fx-target` on links and forms: fetch the destination, swap those selectors.
- `fx-loading-target`, applying the `fx-loading` class while a request is in
  flight.
- `fx-hungry`, for elements that should be swapped on every navigation.
- `<meta name="fx-refresh" fx-interval>` for interval polling.
- History integration, including restoring swapped fragments on `popstate` and
  resetting scroll position on navigation.
- Fallbacks to ordinary navigation on timeout or error, configurable through
  `fx.clickFallback`, `fx.submitFallback` and `fx.historyFallback`.
- `fx.dev.js`, an optional development companion: logging, plus warnings for
  duplicate ids, hungry elements without an id, and targets matching nothing.

### The protocol

- Every request carries an `FX-Target` header naming the fragments the browser
  is about to use, hungry selectors included, so a handler can skip work nobody
  asked for. Ignoring it is correct, just slower.

### The Go package

- `fx.Handler()` serves the script from your binary.
- `fx.Wants(r, selector)` reports whether a request needs a given fragment,
  and is true on any ordinary page load.
