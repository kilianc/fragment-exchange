package web

import (
	"fmt"
	"html"
	"strings"

	g "maragu.dev/gomponents"
)

// Diagrams for the patterns page.
//
// Each one draws a single pattern, in the same idiom as the hook: boxes are
// white on the page's ground, anything fx adds is in the accent colour, and
// nothing is on a grid. They share the p-* class names in style.go.

// pbox is a labelled rounded box.
func pbox(x, y, w, h int, cls, cap string) string {
	s := fmt.Sprintf(`<rect class="%s" x="%d" y="%d" width="%d" height="%d" rx="10" />`, cls, x, y, w, h)
	if cap != "" {
		s += fmt.Sprintf(`<text class="p-cap" x="%d" y="%d">%s</text>`, x+14, y+22, html.EscapeString(cap))
	}
	return s
}

func ptext(x, y int, cls, s string) string {
	return fmt.Sprintf(`<text class="%s" x="%d" y="%d">%s</text>`, cls, x, y, html.EscapeString(s))
}

func pmid(x, y int, cls, s string) string {
	return fmt.Sprintf(`<text class="%s" x="%d" y="%d" text-anchor="middle">%s</text>`, cls, x, y, html.EscapeString(s))
}

func psvg(w, h int, label string, parts ...string) g.Node {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="diagram pattern" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		w, h, html.EscapeString(label))
	b.WriteString(defs())
	for _, p := range parts {
		b.WriteString(p)
	}
	b.WriteString(`</svg>`)
	return g.Raw(b.String())
}

// LayoutDiagram — one document, one hole, and the two hungry pieces outside it.
func LayoutDiagram() g.Node {
	var b strings.Builder

	b.WriteString(pbox(20, 30, 620, 320, "p-doc", ""))
	b.WriteString(ptext(38, 54, "p-doclabel", "the document your server renders, every time"))

	// The hungry pieces: outside #content, updated anyway.
	b.WriteString(pbox(38, 68, 584, 40, "p-hungry", ""))
	b.WriteString(ptext(52, 93, "p-tag", "<title id=\"page-title\" fx-hungry>"))
	b.WriteString(ptext(560, 93, "p-hungrytag", "hungry"))

	b.WriteString(pbox(38, 120, 584, 46, "p-hungry", ""))
	b.WriteString(ptext(52, 148, "p-tag", "<nav id=\"primary-nav\" fx-hungry>"))
	b.WriteString(ptext(560, 148, "p-hungrytag", "hungry"))

	// The hole.
	b.WriteString(pbox(38, 180, 584, 150, "p-hole", ""))
	b.WriteString(ptext(52, 208, "p-tag p-tag-hot", "<main id=\"content\">"))
	b.WriteString(ptext(52, 236, "p-note", "the page. this is what every link asks for."))
	b.WriteString(pbox(52, 250, 556, 26, "p-ghost", ""))
	b.WriteString(pbox(52, 284, 380, 26, "p-ghost", ""))

	// What a link says, and what comes back changed.
	b.WriteString(pbox(670, 110, 310, 160, "p-box", "ONE LINK, ONE ATTRIBUTE"))
	b.WriteString(ptext(686, 156, "p-code", `<a href="/jobs"`))
	b.WriteString(ptext(686, 178, "p-code p-code-hot", `   fx-target="#content">`))
	b.WriteString(ptext(686, 212, "p-note", "It names one thing."))
	b.WriteString(ptext(686, 234, "p-note", "Three get swapped, because"))
	b.WriteString(ptext(686, 254, "p-note", "the server said they were hungry."))

	b.WriteString(arrow(664, 190, 646, 190, "p-line p-line-fx"))

	return psvg(1000, 366,
		"A document with a hungry title and nav outside the content element, and a link that targets only the content",
		b.String())
}

// URLStateDiagram — a form is a URL is a page.
func URLStateDiagram() g.Node {
	var b strings.Builder

	// The form.
	b.WriteString(pbox(20, 40, 250, 150, "p-box", "A PLAIN GET FORM"))
	b.WriteString(pbox(38, 76, 214, 30, "p-field", ""))
	b.WriteString(ptext(50, 96, "p-fieldt", "timeout"))
	b.WriteString(pbox(38, 116, 100, 30, "p-field", ""))
	b.WriteString(ptext(50, 136, "p-fieldt", "open"))
	b.WriteString(pbox(150, 116, 102, 30, "p-submit", ""))
	b.WriteString(pmid(201, 136, "p-submitt", "Filter"))

	b.WriteString(arrow(278, 115, 336, 115, "p-line"))
	b.WriteString(pmid(307, 100, "p-hint", "submit"))

	// Which is just a URL.
	b.WriteString(pbox(344, 60, 320, 110, "p-box", "WHICH IS JUST A URL"))
	b.WriteString(pbox(360, 96, 288, 32, "p-url", ""))
	b.WriteString(ptext(372, 117, "p-urlt", "/jobs?q=timeout&status=open"))
	b.WriteString(ptext(360, 152, "p-note", "Refresh it. Bookmark it. Paste it."))

	b.WriteString(arrow(672, 115, 730, 115, "p-line"))
	b.WriteString(pmid(701, 100, "p-hint", "GET"))

	// Which the server turns back into a page.
	b.WriteString(pbox(738, 40, 242, 150, "p-box", "WHICH THE SERVER READS"))
	b.WriteString(ptext(754, 96, "p-code", "f := parseFilters(r)"))
	b.WriteString(ptext(754, 120, "p-code", "rows := search(f)"))
	b.WriteString(ptext(754, 158, "p-note", "No state anywhere else."))

	b.WriteString(pmid(500, 216, "p-caption", "The form makes the URL. The URL makes the page. There is no third copy."))

	return psvg(1000, 236,
		"A GET form produces a URL, and the server turns that URL back into a page",
		b.String())
}

// DetailPaneDiagram — open and closed are two URLs, so back closes it.
func DetailPaneDiagram() g.Node {
	var b strings.Builder

	closed := 30
	open := 530

	// Closed.
	b.WriteString(pbox(closed, 44, 420, 210, "p-box", ""))
	b.WriteString(pbox(closed+16, 60, 388, 30, "p-url", ""))
	b.WriteString(ptext(closed+28, 80, "p-urlt", "/jobs?q=timeout"))
	b.WriteString(pbox(closed+16, 104, 388, 134, "p-ghost", ""))
	b.WriteString(ptext(closed+30, 130, "p-note", "the list, and no pane"))

	// Open.
	b.WriteString(pbox(open, 44, 420, 210, "p-box", ""))
	b.WriteString(pbox(open+16, 60, 388, 30, "p-url p-url-hot", ""))
	b.WriteString(ptext(open+28, 80, "p-urlt p-urlt-hot", "/jobs?q=timeout&open=1421"))
	b.WriteString(pbox(open+16, 104, 240, 134, "p-ghost", ""))
	b.WriteString(pbox(open+266, 104, 138, 134, "p-pane", ""))
	b.WriteString(ptext(open+278, 130, "p-panet", "#detail"))
	b.WriteString(ptext(open+278, 152, "p-note", "job 1421"))

	// There and back.
	b.WriteString(arrow(462, 108, 518, 108, "p-line p-line-fx"))
	b.WriteString(pmid(490, 96, "p-hint p-hint-fx", "open"))
	b.WriteString(arrow(518, 190, 462, 190, "p-line"))
	b.WriteString(pmid(490, 214, "p-hint", "back"))

	b.WriteString(pmid(500, 292, "p-caption",
		"Closing the pane is a URL you have already been to, so the back button does it for free."))

	return psvg(1000, 312,
		"The same page with and without a detail pane, differing only by an open query parameter",
		b.String())
}

// PollingDiagram — the timer lives in the fragment and dies with it.
func PollingDiagram() g.Node {
	var b strings.Builder

	// On the page.
	b.WriteString(pbox(20, 44, 430, 240, "p-box", "WHILE YOU ARE ON THE PAGE"))
	b.WriteString(pbox(38, 84, 394, 44, "p-hungry", ""))
	b.WriteString(ptext(52, 111, "p-tag p-tag-hot", `<meta name="fx-refresh" fx-interval="5000">`))

	b.WriteString(pbox(38, 142, 394, 124, "p-hole", ""))
	b.WriteString(ptext(52, 168, "p-tag", `<div id="jobs">`))
	for i := 0; i < 3; i++ {
		b.WriteString(pbox(52, 182+i*26, 366, 18, "p-ghost", ""))
	}

	// The loop.
	b.WriteString(`<path class="p-line p-line-fx" d="M462 130 C520 130 520 200 462 200" fill="none" />`)
	b.WriteString(`<use href="#fx-arrow-r" class="p-head p-head-fx" x="462" y="196" transform="rotate(180 469 200)" />`)
	b.WriteString(ptext(500, 168, "p-hint p-hint-fx", "every 5s"))

	// After you leave.
	b.WriteString(pbox(570, 44, 410, 240, "p-box", "AFTER YOU NAVIGATE AWAY"))
	b.WriteString(pbox(588, 84, 374, 44, "p-gone", ""))
	b.WriteString(pmid(775, 111, "p-gonet", "the meta tag was swapped out with the fragment"))

	b.WriteString(pbox(588, 142, 374, 124, "p-hole", ""))
	b.WriteString(ptext(602, 168, "p-tag", `<div id="something-else">`))
	b.WriteString(ptext(602, 198, "p-note", "The timer is cancelled."))
	b.WriteString(ptext(602, 220, "p-note", "Nothing to unsubscribe from,"))
	b.WriteString(ptext(602, 240, "p-note", "because there is no lifecycle."))

	b.WriteString(pmid(500, 320, "p-caption",
		"The tag lives inside the fragment it refreshes, so leaving the page removes the timer with it."))

	return psvg(1000, 340,
		"A polling meta tag inside a fragment, and the same page after navigation with the timer gone",
		b.String())
}

// MutationDiagram — post, redirect, get, with fx following along.
func MutationDiagram() g.Node {
	var b strings.Builder

	steps := []struct {
		X          int
		Cap        string
		Line1      string
		Line2      string
		Hot        bool
		Annotation string
	}{
		{20, "1 · THE FORM POSTS", "POST /jobs/1421/requeue", "fx-target=\"#jobs, #detail\"", true, "A form, doing what forms do."},
		{350, "2 · THE SERVER ANSWERS", "303 See Other", "Location: /jobs?open=1421", false, "Not a fragment. A place to go."},
		{680, "3 · FX FOLLOWS IT", "GET /jobs?open=1421", "FX-Target: #jobs, #detail", true, "And puts that URL in the bar."},
	}

	for _, s := range steps {
		b.WriteString(pbox(s.X, 40, 300, 150, "p-box", s.Cap))
		cls := "p-code"
		if s.Hot {
			cls = "p-code p-code-hot"
		}
		b.WriteString(ptext(s.X+16, 88, cls, s.Line1))
		b.WriteString(ptext(s.X+16, 112, "p-code", s.Line2))
		b.WriteString(ptext(s.X+16, 156, "p-note", s.Annotation))
	}

	b.WriteString(arrow(328, 115, 342, 115, "p-line"))
	b.WriteString(arrow(658, 115, 672, 115, "p-line p-line-fx"))

	b.WriteString(pmid(500, 232, "p-caption",
		"The browser ends on an ordinary GET of an ordinary page, so a refresh re-reads instead of re-submitting."))

	return psvg(1000, 252,
		"A POST answered with a 303 redirect, which fx follows to a normal GET",
		b.String())
}

// CheapDiagram — the same handler, asked two different ways.
func CheapDiagram() g.Node {
	var b strings.Builder

	frags := []struct {
		Name string
		Cost string
	}{
		{"#primary-nav", "2ms"},
		{"#summary", "900ms"},
		{"#jobs", "40ms"},
		{"#detail", "8ms"},
	}

	col := func(x int, cap, header string, hot bool, want map[string]bool, total, note string) string {
		var c strings.Builder
		c.WriteString(pbox(x, 40, 440, 250, "p-box", cap))

		cls := "p-code"
		if hot {
			cls = "p-code p-code-hot"
		}
		c.WriteString(ptext(x+16, 86, cls, header))

		for i, f := range frags {
			y := 108 + i*34
			on := want[f.Name]
			bar, txt, state := "p-run", "p-runt", "rendered"
			if !on {
				bar, txt, state = "p-skip", "p-skipt", "skipped"
			}
			c.WriteString(fmt.Sprintf(`<rect class="%s" x="%d" y="%d" width="290" height="24" rx="5" />`, bar, x+16, y))
			c.WriteString(ptext(x+28, y+17, txt, f.Name))
			c.WriteString(fmt.Sprintf(`<text class="%s" x="%d" y="%d" text-anchor="end">%s</text>`, txt, x+296, y+17, f.Cost))
			c.WriteString(ptext(x+318, y+17, "p-state "+bar+"t", state))
		}

		c.WriteString(ptext(x+16, 268, "p-total", total))
		c.WriteString(ptext(x+180, 268, "p-note", note))
		return c.String()
	}

	all := map[string]bool{"#primary-nav": true, "#summary": true, "#jobs": true, "#detail": true}
	poll := map[string]bool{"#jobs": true, "#detail": true, "#primary-nav": true}

	b.WriteString(col(20, "A PAGE LOAD", "FX-Target: (none)", false, all, "950ms", "renders everything, always"))
	b.WriteString(col(540, "A POLL, EVERY 5 SECONDS", "FX-Target: #jobs, #detail, #primary-nav", true, poll, "50ms", "skips only the roll-up"))

	b.WriteString(pmid(500, 330, "p-caption",
		"One handler. fx.Wants is true whenever the header is absent, so the page load still gets the whole page."))

	return psvg(1000, 350,
		"The same handler rendering every fragment for a page load, and only two fragments for a poll",
		b.String())
}
