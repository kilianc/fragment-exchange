// Package fx embeds the fx client library and implements the server half of
// its protocol.
//
// fx is a JavaScript file. This package exists so a Go server can serve it out
// of its own binary, and so handlers can answer the one question the protocol
// asks: which fragments does the client actually want?
//
//	//go:embed nothing — fx.js is already in here
//	mux.Handle("/fx.js", fx.Handler())
//
//	func page(w http.ResponseWriter, r *http.Request) {
//		var rows []Row
//		if fx.Wants(r, "#results") {
//			rows = expensiveQuery()   // skipped when the client wants only the nav
//		}
//		render(w, rows)
//	}
package fx

import (
	_ "embed"
	"net/http"
	"strings"
)

// TargetHeader carries the selectors the client is about to swap. It is empty
// on an ordinary page load.
const TargetHeader = "FX-Target"

//go:embed fx.js
var script []byte

//go:embed fx.dev.js
var devScript []byte

// Script is the library: serve it, or copy it into your project and forget
// this package exists.
func Script() []byte { return script }

// DevScript is the development companion. It turns on logging and warns about
// documents fx cannot swap correctly. Do not serve it in production.
func DevScript() []byte { return devScript }

// Version is the version of the embedded library.
const Version = "1.1.3"

// Handler serves fx.js and fx.dev.js at whatever paths it is mounted on,
// keyed by the last path segment.
//
//	mux.Handle("/fx.js", fx.Handler())
//	mux.Handle("/fx.dev.js", fx.Handler())
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := script
		if strings.HasSuffix(r.URL.Path, "fx.dev.js") {
			body = devScript
		}

		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(body)
	})
}

// Targets returns the selectors the client asked for, or nil on an ordinary
// page load.
func Targets(r *http.Request) []string {
	raw := r.Header.Get(TargetHeader)
	if raw == "" {
		return nil
	}

	var targets []string
	for _, s := range strings.Split(raw, ",") {
		if s = strings.TrimSpace(s); s != "" {
			targets = append(targets, s)
		}
	}
	return targets
}

// IsFragment reports whether this request came from fx rather than from the
// address bar.
//
// Use it to decide what a response should *contain* — a modal that is already
// open does not need its opening animation replayed. Do not use it to decide
// whether to render at all: fx may still ask for the whole document.
func IsFragment(r *http.Request) bool {
	return r.Header.Get(TargetHeader) != ""
}

// Wants reports whether a fragment has to be rendered for this request.
//
// It is true for every ordinary page load, and true for a fragment request
// that named this selector. That default is the safe one: a handler that has
// never heard of fx keeps working, and one that opts in only skips work the
// client has said it will throw away.
//
// A selector matches if the client asked for it exactly, or if either one
// contains the other. "#content" covers "#content .row", because swapping the
// container swaps the row with it; and a client that asked for "#content .row"
// needs "#content" rendered, because that is the only way the response can
// contain the row.
func Wants(r *http.Request, selector string) bool {
	targets := Targets(r)
	if targets == nil {
		return true
	}

	for _, t := range targets {
		if t == selector || isInside(selector, t) || isInside(t, selector) {
			return true
		}
	}
	return false
}

// isInside reports whether inner names something within outer: the same
// selector with more on the end, joined by a descendant or child combinator.
//
// Sibling combinators are deliberately not here. "#content + .row" is beside
// the container, not in it, so rendering "#content" does not produce it.
func isInside(inner, outer string) bool {
	rest, found := strings.CutPrefix(inner, outer)
	if !found || rest == "" {
		return false
	}

	// A combinator has to separate the two, or outer merely spells the start of
	// inner: "#content" and "#contentious".
	trimmed := strings.TrimLeft(rest, " \t\n")
	if trimmed == rest && !strings.HasPrefix(rest, ">") {
		return false
	}

	return trimmed != "" && !strings.ContainsAny(trimmed[:1], "+~")
}

// WantsAny reports whether any of the selectors has to be rendered.
func WantsAny(r *http.Request, selectors ...string) bool {
	for _, s := range selectors {
		if Wants(r, s) {
			return true
		}
	}
	return false
}
