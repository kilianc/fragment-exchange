# fx.js — Fragment eXchange

A minimal JavaScript library for fragment-oriented navigation in server-rendered applications.

## Overview

fx enables partial page updates without changing the foundational architecture of your application. The server renders complete HTML documents; fx swaps the parts that changed.

```html
<script src="/fx.js" defer></script>

<a href="/reports" fx-target="#content">Reports</a>
<main id="content"><!-- Server-rendered HTML --></main>
```

If JavaScript is unavailable, the link functions normally.

## Motivation

Modern web development often assumes that client-side frameworks are the default choice, even for applications whose data flow and operational requirements are inherently server-centric. Engineering teams frequently carry the operational burden of front-end toolchains, dependency ecosystems, and dual-state architectures—despite not benefiting from the abstractions those systems were designed to provide.

fx explores the opposite direction. Rather than replacing or overshadowing server-rendered HTML, fx provides a minimal mechanism for fragment updates that remains strictly subordinate to server output.

## Core Model

1. The server renders a complete HTML document.
2. A navigation event triggers a request for the next document.
3. The returned document is parsed in memory.
4. The fragment associated with `fx-target` is extracted and swapped into the current DOM.
5. Browser history is updated.
6. Failure modes fall back to normal navigation.

This approach avoids virtual DOMs, hydration, and client-side state.

## Architectural Principles

- **Server authority**: All rendering originates from the backend.
- **URL-driven state**: The URL encodes the entire view state.
- **Minimal API surface**: One primary affordance.
- **Failure safety**: fx declines into ordinary navigation when needed.

## API Reference

| Attribute | Description |
|-----------|-------------|
| `fx-target` | CSS selector(s) identifying fragments to replace |
| `fx-loading-target` | Elements that receive `fx-loading` class during requests |
| `fx-hungry` | Marks element for inclusion in all swaps (requires `id`) |
| `fx-interval` | On `<meta name="fx-refresh">`, specifies polling interval in ms |

### Runtime Configuration

- `fx.timeout` — Request timeout (default: 10000ms)
- `fx.clickFallback` — Handler for failed link navigations
- `fx.submitFallback` — Handler for failed form submissions
- `fx.historyFallback` — Handler for failed history navigations

### Server Protocol

Each request includes an `FX-Target` header containing the requested selectors. The server may return a complete document or only the requested fragments.

## Non-Goals

fx does not provide components, client-side routing, reactivity, global state management, or template systems. These omissions are deliberate.

## Installation

Copy `fx.js` into your project. No build step required.

```text
fx.js      — Core library (~300 lines)
fx.dev.js  — Development helper (optional)
```

## Documentation

```bash
python -m http.server 8000
open http://localhost:8000/docs/
```

## License

MIT
