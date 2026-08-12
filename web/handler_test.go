package web_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	fx "github.com/kilianc/fragment-exchange"
	"github.com/kilianc/fragment-exchange/web"
)

// get makes a request against the real site handler. target, when set, is the
// FX-Target header — i.e. this is fx asking rather than the address bar.
func get(t *testing.T, path, target string) (*http.Response, string) {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, path, nil)
	if target != "" {
		r.Header.Set(fx.TargetHeader, target)
	}

	w := httptest.NewRecorder()
	web.Handler().ServeHTTP(w, r)

	return w.Result(), w.Body.String()
}

func TestEveryPageAnswers(t *testing.T) {
	for _, p := range web.Pages() {
		t.Run(p.Path(), func(t *testing.T) {
			res, body := get(t, p.Path(), "")

			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", res.StatusCode)
			}
			if got := res.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
				t.Errorf("Content-Type = %q", got)
			}
			if !strings.Contains(body, ">"+p.Title+"<") {
				t.Errorf("the page does not contain its own title %q", p.Title)
			}
		})
	}
}

func TestUnknownPathIs404(t *testing.T) {
	res, body := get(t, "/no-such-page", "")

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", res.StatusCode)
	}
	// Still a whole page, not a bare error string.
	if !strings.Contains(body, "<html") {
		t.Error("the 404 is not a rendered page")
	}
}

// A fragment request still gets a complete document. This is the invariant the
// whole approach rests on: the server never learns to speak in fragments.
func TestFragmentRequestsStillGetAWholeDocument(t *testing.T) {
	for _, target := range []string{"", "#content", "#detail", "#jobs, #detail"} {
		name := target
		if name == "" {
			name = "(page load)"
		}

		t.Run(name, func(t *testing.T) {
			res, body := get(t, "/demo", target)

			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", res.StatusCode)
			}
			for _, want := range []string{"<!doctype html>", "<html", "<head", "<body", `id="content"`} {
				if !strings.Contains(body, want) {
					t.Errorf("response is missing %q", want)
				}
			}
		})
	}
}

// The parts no link asks for but every navigation needs.
func TestHungryElements(t *testing.T) {
	_, body := get(t, "/", "")

	for _, want := range []string{`id="page-title" fx-hungry`, `id="primary-nav" class="primary" fx-hungry`} {
		if !strings.Contains(body, want) {
			t.Errorf("response is missing a hungry element: %q", want)
		}
	}
}

// The figure the home page opens with is a two-state drawing, and the state is
// a query string like any other. It is the smallest possible example of the
// thing the page is arguing for, so it has to hold to the same rules: reachable
// by URL, rendered by the server, swapped by fx.
func TestTheOpeningFigureIsAURL(t *testing.T) {
	_, plain := get(t, "/", "")
	_, enabled := get(t, "/?fx=on", "")

	// The picture makes its point by changing one thing: how much of the page
	// the browser replaces.
	if !strings.Contains(plain, "all of it — replaced") {
		t.Error("/ does not say the whole page is replaced")
	}
	if !strings.Contains(enabled, "just this — replaced") {
		t.Error("/?fx=on does not say only one part is replaced")
	}

	// And by changing nothing else. If the scene is not word for word the same
	// on both sides, the drawing is arguing against the page it opens.
	for _, want := range []string{
		"Every website you have ever used works like this.",
		"YOUR SERVER",
		"a whole page comes back",
		"every time — with fx or without it",
	} {
		if !strings.Contains(plain, want) || !strings.Contains(enabled, want) {
			t.Errorf("%q is not in both states of the figure", want)
		}
	}

	// And the element the toggle asks for has to come back in both states, or
	// fx has nothing to swap and falls back to a reload.
	for _, path := range []string{"/", "/?fx=on"} {
		_, body := get(t, path, "#hook-figure")
		if !strings.Contains(body, `id="hook-figure"`) {
			t.Errorf("%s: the fragment the toggle targets is not in the response", path)
		}
	}
}

// The claim the demo makes in its own log: a request that does not name the
// expensive fragment does not pay for it.
func TestFxTargetSkipsUnrequestedWork(t *testing.T) {
	t.Run("a page load renders everything", func(t *testing.T) {
		start := time.Now()
		_, body := get(t, "/demo", "")
		elapsed := time.Since(start)

		if strings.Contains(body, `id="summary" class="panel" hidden`) {
			t.Error("the roll-up was skipped on a page load; a page load has to render everything")
		}
		if elapsed < 200*time.Millisecond {
			t.Errorf("the page load took %v — the expensive fragment cannot have run", elapsed)
		}
	})

	t.Run("a poll skips the roll-up", func(t *testing.T) {
		start := time.Now()
		_, body := get(t, "/demo", "#jobs, #detail")
		elapsed := time.Since(start)

		if !strings.Contains(body, `id="summary" class="panel" hidden`) {
			t.Error("the roll-up was rendered for a request that did not ask for it")
		}
		if strings.Contains(body, `id="jobs" class="panel" hidden`) {
			t.Error("the jobs table was skipped for a request that did ask for it")
		}
		if elapsed > 150*time.Millisecond {
			t.Errorf("the fragment request took %v — it should not be paying for the roll-up", elapsed)
		}
	})

	t.Run("opening a pane skips everything else", func(t *testing.T) {
		_, body := get(t, "/demo?open=1401", "#detail")

		if !strings.Contains(body, `id="summary" class="panel" hidden`) {
			t.Error("the roll-up was rendered for a #detail request")
		}
		if !strings.Contains(body, `id="jobs" class="panel" hidden`) {
			t.Error("the jobs table was rendered for a #detail request")
		}
		if strings.Contains(body, `id="detail" class="panel" hidden`) {
			t.Error("the detail pane was skipped for a #detail request")
		}
	})
}

// Everything you can see is in the URL, so the URL alone has to produce it.
func TestStateComesFromTheURL(t *testing.T) {
	_, all := get(t, "/demo", "#jobs")
	if !strings.Contains(all, "terraform-plan") {
		t.Fatal("the unfiltered table is missing a job it should contain")
	}

	_, filtered := get(t, "/demo?q=terraform", "#jobs")
	if !strings.Contains(filtered, "terraform-plan") {
		t.Error("the filtered table dropped the job that matches")
	}
	if strings.Contains(filtered, "link-check") {
		t.Error("the filtered table kept a job that does not match")
	}

	_, empty := get(t, "/demo?q=nothing-matches-this", "#jobs")
	if !strings.Contains(empty, "Nothing matches that filter") {
		t.Error("an empty result set has no empty state")
	}
}

func TestDetailPaneOpensFromTheQueryString(t *testing.T) {
	_, closed := get(t, "/demo", "#detail")
	if !strings.Contains(closed, "Open a job to see it here") {
		t.Error("the pane is not closed without ?open=")
	}

	_, open := get(t, "/demo?open=1405", "#detail")
	if !strings.Contains(open, "migrate-staging") {
		t.Error("?open=1405 did not render that job")
	}
	if strings.Contains(open, "Open a job to see it here") {
		t.Error("the pane is still closed with ?open=")
	}
}

// A mutation answers with a redirect, so the browser lands on an ordinary GET.
func TestRequeueRedirects(t *testing.T) {
	form := strings.NewReader("id=1403&q=build&status=running")
	r := httptest.NewRequest(http.MethodPost, "/demo/requeue", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	web.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", w.Code)
	}

	location := w.Header().Get("Location")
	for _, want := range []string{"open=1403", "q=build", "status=running"} {
		if !strings.Contains(location, want) {
			t.Errorf("Location %q does not carry %q — the redirect loses state", location, want)
		}
	}
}

// The site serves the library it documents, from the binary.
func TestLibraryIsServed(t *testing.T) {
	res, body := get(t, "/fx.js", "")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if !strings.Contains(body, "window.fx") {
		t.Error("/fx.js is not the library")
	}
	if !strings.Contains(body, fx.Version) {
		t.Errorf("/fx.js does not report version %s", fx.Version)
	}
}

// A GSX call written in markup position without braces is valid GSX: it
// compiles to a text node, and the page ships with Go source printed on it.
// Nothing else catches that, so this does.
func TestNoUnsplicedGoLeakedIntoThePage(t *testing.T) {
	markers := []string{
		`Text(&#34;`, `Text("`,
		`Group(`, `Section(&#34;`, `Snippet(&#34;`, `Pull(`, `Note(&#34;`, `Attr(&#34;`,
	}

	for _, p := range web.Pages() {
		_, body := get(t, p.Path(), "")

		for _, m := range markers {
			if strings.Contains(body, m) {
				t.Errorf("%s: %q appears in the rendered page — a GSX call is missing its braces", p.Path(), m)
			}
		}
	}
}

// Every link fx handles has to be a link a browser can follow on its own.
func TestEveryFxLinkHasARealHref(t *testing.T) {
	for _, p := range web.Pages() {
		_, body := get(t, p.Path(), "")

		for _, chunk := range strings.Split(body, "<a ")[1:] {
			tag, _, _ := strings.Cut(chunk, ">")
			if !strings.Contains(tag, "fx-target") {
				continue
			}
			// The sidebar-style anchors with no href are not navigation.
			if !strings.Contains(tag, "href=") {
				t.Errorf("%s: a link with fx-target has no href: <a %s>", p.Path(), tag)
			}
		}
	}
}

// GSX follows JSX's whitespace rules: a newline between prose and a tag is
// removed, not collapsed to a space. Wrapping a paragraph so that an inline
// <code> lands at the start of a line therefore glues two words together, and
// it is invisible in the source.
func TestNoGluedWordsAroundInlineCode(t *testing.T) {
	glued := regexp.MustCompile(`[A-Za-z0-9,;:)]<code>|</code>[A-Za-z0-9]`)

	for _, p := range web.Pages() {
		_, body := get(t, p.Path(), "")

		// Code samples are highlighted into spans, not <code> runs, so the
		// only <code> elements left are the inline ones in prose.
		for _, m := range glued.FindAllString(body, -1) {
			t.Errorf("%s: %q — a space was lost between prose and inline code", p.Path(), m)
		}
	}
}
