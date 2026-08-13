# Changelog

Releases are git tags. Each one is served from jsDelivr at
`cdn.jsdelivr.net/gh/kilianc/fragment-exchange@<tag>/fx.js`, immutably.

fx follows semver over three surfaces: the attributes, the `fx` object, and the
`FX-Target` header. See the versioning notes in the README for why the header
counts.

## 1.1.4

### The library

- An empty selector in an `fx-target` list is ignored. A stray comma produced a
  selector `querySelector` rejects, and that threw before anything was swapped —
  so the whole navigation failed into the browser and the link quietly became an
  ordinary page load. The server already dropped those entries when it read the
  header back, and so did `fx.dev.js` when it checked the attribute, which is
  why nothing warned about it. All three now agree.

### Inside

- The hungry-element walk, which ran once over the current page and once over
  the response, is one function. No behaviour change.

## 1.1.3

### The library

- A submission the server does not redirect leaves history alone. fx pushed the
  address the form posted to, then re-fetched it with `GET` when the user came
  back — a request the server never agreed to answer, and one no browser makes:
  going back to a submitted page replays nothing and asks the network for
  nothing, because the browser kept the response. fx keeps nothing, so it no
  longer claims a URL it could only restore incorrectly.
- Post/redirect/get is unchanged and is still the way to give a submission a URL
  of its own. The redirect is the server naming an address that answers a `GET`,
  and it earns the history entry the submission itself cannot.

## 1.1.2

### The library

- A link to another origin is left to the browser. fx took the click, then the
  history API refused the foreign URL and threw before the request was made —
  cancelling the navigation without performing one, so the link did nothing at
  all. `mailto:` and `tel:` links carrying an `fx-target` were dead the same way.
- Forms are read through their attributes, not their properties. A form exposes
  its controls as properties and they win over the ones the platform defines, so
  a field named `action` or `method` replaced the form's own — and a button
  named `submit`, which is ordinary markup, shadowed `form.submit()` and made
  the fallback throw, losing the submission it existed to rescue.
- A poll that times out tries again. `fx-refresh` shared one `AbortController`
  across every tick, and a signal cannot be un-aborted, so the timeout meant to
  cancel one slow request stopped the polling with it — permanently, and without
  saying anything. Each request now carries its own.
- An abort the browser made is treated as a cancel, not a failure. The browser
  drops every in-flight fetch when a real navigation starts, and fx answered
  that rejection with its fallback — replacing the location, and so overriding
  the navigation the user had just asked for.
- A poller set up by a redirected response polls where the redirect landed. The
  timers were started before the redirect was recorded, so they read the address
  bar a moment too early and kept the URL the server had just sent them away
  from — for the life of the page.
- A `GET` form replaces the query string in its `action` instead of appending to
  it, which is what a browser does with the same form. `action="/search?tab=all"`
  produced `/search?tab=all?q=hello`; if the action also carried a fragment, the
  fields were appended to that and never reached the server at all.
- Going back to a history entry fx never wrote leaves it alone. An in-page
  anchor makes one, and so does an application recording its own state; fx used
  to fetch the whole page for these and then swap nothing out of it, because
  nothing on the entry named what to swap. The document they belong to is the
  one already on screen.
- A click or submit another listener has already cancelled is left alone. fx
  delegates from the document, so it hears about an event last and was
  overruling code that had said no — a declined confirmation, or a control that
  handles its own click.
- `fx-refresh` starts polling when `fx.js` is loaded after the document. fx
  waited on `DOMContentLoaded` without checking whether it had already fired, so
  loading the library with `async`, or from an injected tag, meant polling never
  began. `fx.dev.js` already made this check.
- A failed response that rendered the requested fragments is swapped in. Any
  status other than 2xx was discarded and the navigation handed to the browser,
  which for a form meant posting it a second time — repeating whatever the first
  attempt had already done. A rejected form coming back with its errors is the
  ordinary answer to a `422`, not a failure. A failed response that does not
  contain the fragments is still an error page fx cannot use, and still ends in
  an ordinary navigation.

### The Go package

- `fx.Wants` answers for a container whose contents were the target. A client
  asking for `#content .row` left `fx.Wants(r, "#content")` false, so the
  handler skipped the only thing that could have carried the row, and the swap
  found nothing in the response — a click that quietly did nothing. It now
  matches when either selector contains the other.
- `fx.Wants` no longer counts a sibling as inside. `#content + .row` reads as a
  descendant to a prefix comparison; it names an element beside the container,
  which rendering the container does not produce.

## 1.1.1

### The library

- History entries keep the state the application put on them. fx writes its own
  bookkeeping onto an entry before navigating away, and that write replaced the
  entry's state wholesale, silently discarding anything the page had stored with
  its own `pushState` or `replaceState`.
- A failed navigation no longer clears the entry's state when it restores the
  URL, and no longer touches history at all when the URL never moved — a
  navigation to the address already in the address bar left nothing to undo.

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
