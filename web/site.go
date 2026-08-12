// Package web is fx.ciuffolo.com: the argument for fx, the reference, and a
// working application that uses it.
//
// The site is server-rendered Go. Every page is a complete document, every
// link works with JavaScript turned off, and fx swaps <main> when it is on —
// which is the entire thesis, so it would be embarrassing to do otherwise.
package web

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	fx "github.com/kilianc/fragment-exchange"
	g "maragu.dev/gomponents"
)

// fxVersion is the version of the library this site documents, read from the
// library itself so the reference cannot quote a number that does not exist.
const fxVersion = fx.Version

// Page is one page of the site.
type Page struct {
	Slug  string // "" is the index
	Nav   string // label in the top nav; empty keeps it out
	Title string // the <title>, and the heading unless Headline is set
	// Headline is the <h1>, when the thing to say to a reader is not the thing
	// to put in a browser tab.
	Headline string
	Lede     string
	Wide     bool
	Body     func(c *Ctx) g.Node
}

// Path is the URL this page is served from.
func (p Page) Path() string {
	if p.Slug == "" {
		return "/"
	}
	return "/" + p.Slug
}

// Ctx is what a page body gets: the request, and what fx asked for.
type Ctx struct {
	R    *http.Request
	Page Page

	// Fragment is true when fx made this request rather than the address bar.
	Fragment bool
	// Targets are the selectors fx is about to swap, empty on a page load.
	Targets []string

	log   *sessionLog
	entry *logEntry
	start time.Time
}

// Wants reports whether a fragment has to be rendered for this request.
// A page load wants everything.
func (c *Ctx) Wants(selector string) bool { return fx.Wants(c.R, selector) }

// Pages returns every page, in nav order.
func Pages() []Page {
	return []Page{
		IndexPage(),
		PatternsPage(),
		ReferencePage(),
		DemoPage(),
	}
}

func pageFor(slug string) (Page, bool) {
	for _, p := range Pages() {
		if p.Slug == slug {
			return p, true
		}
	}
	return Page{}, false
}

// Handler is the whole site.
func Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /fx.js", fx.Handler())
	mux.Handle("GET /fx.dev.js", fx.Handler())
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
	})

	// The demo's one mutation. It answers with a redirect, which is what a
	// form post should do — and fx follows it and rewrites the address bar.
	mux.HandleFunc("POST /demo/requeue", requeueHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.Trim(r.URL.Path, "/")

		page, ok := pageFor(slug)
		if !ok {
			serve(w, r, notFoundPage(), http.StatusNotFound)
			return
		}

		serve(w, r, page, http.StatusOK)
	})

	return mux
}

func serve(w http.ResponseWriter, r *http.Request, page Page, status int) {
	s := sessionFrom(w, r)
	log := logFor(s)

	c := &Ctx{
		R:        r,
		Page:     page,
		Fragment: fx.IsFragment(r),
		Targets:  fx.Targets(r),
		log:      log,
		start:    s.Start,
	}

	// The log panel is part of the page it describes, so the entry for this
	// request has to exist before the page renders.
	c.entry = log.record(r, page)

	var buf bytes.Buffer
	if err := Layout(c).Render(&buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func notFoundPage() Page {
	return Page{
		Slug:  "404",
		Title: "Not found",
		Lede:  "That page does not exist. The ones that do are in the nav.",
		Body:  func(c *Ctx) g.Node { return g.Text("") },
	}
}
