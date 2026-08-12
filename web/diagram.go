package web

import (
	"fmt"
	"html"
	"strings"

	g "maragu.dev/gomponents"
)

// The lifecycle diagram.
//
// Three rows, one URL, one handler. Each row adds something and each row is a
// place you are allowed to stop — which is the argument the picture has to
// make, because it is the one people do not believe until they see it laid
// out: the middle row costs the backend nothing at all.
//
// It is inline SVG with CSS variables for colour, so it follows the light and
// dark themes for free and adds no request.

type frag struct {
	Name     string
	Cost     string
	Computed bool
	Swapped  bool
}

// LifecycleDiagram returns the whole figure.
func LifecycleDiagram() g.Node {
	shared := []frag{
		{Name: "#primary-nav", Cost: "2ms"},
		{Name: "#summary", Cost: "900ms"},
		{Name: "#results", Cost: "120ms"},
		{Name: "#detail", Cost: "8ms"},
	}

	all := func(swapped string) []frag {
		out := make([]frag, len(shared))
		for i, f := range shared {
			f.Computed = true
			f.Swapped = f.Name == swapped
			out[i] = f
		}
		return out
	}

	onlyDetail := all("#detail")
	for i := range onlyDetail {
		onlyDetail[i].Computed = onlyDetail[i].Name == "#detail"
	}

	var b strings.Builder

	fmt.Fprintf(&b, `<svg class="diagram" viewBox="0 0 1080 606" role="img" `+
		`aria-label="The fx lifecycle: the same URL and the same handler at three levels of adoption">`)

	b.WriteString(defs())
	b.WriteString(spine())

	b.WriteString(row(rowSpec{
		Y:        76,
		Index:    "0",
		Title:    "A website",
		Subtitle: "No fx. No JavaScript at all, if you like.",
		Request:  []string{"GET /reports?status=open&amp;open=1421"},
		Response: "200 · a complete HTML document",
		Frags:    all(""),
		WholeDoc: true,
		Outcome:  "Works. Always has.",
		Detail:   "The browser throws the old document away and paints a new one.",
	}))

	b.WriteString(row(rowSpec{
		Y:        282,
		Index:    "1",
		Title:    "Add one attribute",
		Subtitle: `The link gains fx-target="#detail". Nothing on the server changes.`,
		Request:  []string{"GET /reports?status=open&amp;open=1421", "FX-Target: #detail"},
		Response: "200 · a complete HTML document",
		Frags:    all("#detail"),
		Outcome:  "Zero backend cost.",
		Detail:   "The handler is untouched and still renders everything. One element is swapped instead of the page — no flash, no scroll jump, and the URL is still the URL.",
	}))

	b.WriteString(row(rowSpec{
		Y:        488,
		Index:    "2",
		Title:    "Answer the header",
		Subtitle: "The handler reads FX-Target and skips what nobody asked for.",
		Request:  []string{"GET /reports?status=open&amp;open=1421", "FX-Target: #detail"},
		Response: "200 · only what was asked for",
		Frags:    onlyDetail,
		Outcome:  "1,022ms → 8ms",
		Detail:   "Worth doing once a fragment is expensive enough to be a component of its own — a modal, a widget, a roll-up.",
		Fast:     true,
	}))

	b.WriteString(`</svg>`)

	return g.Raw(b.String())
}

func defs() string {
	return `<defs>
    <path id="fx-arrow-r" d="M0 0 l7 4 l-7 4 z" />
    <path id="fx-arrow-l" d="M7 0 l-7 4 l7 4 z" />
  </defs>`
}

// spine is the header band: the one thing all three rows have in common.
func spine() string {
	return `<g>
    <text class="d-kicker" x="28" y="20">EVERY SCREEN IS A URL</text>
    <rect class="d-url" x="28" y="30" width="440" height="28" rx="7" />
    <text class="d-mono" x="44" y="49">/reports?status=open&amp;open=1421</text>
    <text class="d-note" x="486" y="49">One address. One handler. Three ways to ask for it.</text>
  </g>`
}

type rowSpec struct {
	Y        int
	Index    string
	Title    string
	Subtitle string
	Request  []string
	Response string
	Frags    []frag
	WholeDoc bool
	Outcome  string
	Detail   string
	Fast     bool
}

func row(s rowSpec) string {
	var b strings.Builder
	y := s.Y

	boxY := y + 46
	boxH := 122

	fmt.Fprintf(&b, `<g>`)

	// Heading.
	fmt.Fprintf(&b, `<circle class="d-badge" cx="40" cy="%d" r="13" />`, y+10)
	fmt.Fprintf(&b, `<text class="d-badge-n" x="40" y="%d">%s</text>`, y+15, s.Index)
	fmt.Fprintf(&b, `<text class="d-title" x="62" y="%d">%s</text>`, y+15, html.EscapeString(s.Title))
	fmt.Fprintf(&b, `<text class="d-sub" x="62" y="%d">%s</text>`, y+33, s.Subtitle)

	// Browser: the page, as four stacked fragments.
	browserClass := "d-box"
	if s.WholeDoc {
		browserClass = "d-box d-box-hot"
	}
	fmt.Fprintf(&b, `<rect class="%s" x="28" y="%d" width="200" height="%d" rx="9" />`, browserClass, boxY, boxH)
	fmt.Fprintf(&b, `<text class="d-cap" x="40" y="%d">BROWSER</text>`, boxY+17)

	for i, f := range s.Frags {
		barY := boxY + 26 + i*23
		cls := "d-frag"
		if f.Swapped {
			cls = "d-frag d-frag-swap"
		}
		fmt.Fprintf(&b, `<rect class="%s" x="42" y="%d" width="172" height="18" rx="4" />`, cls, barY)

		label := "d-frag-t"
		if f.Swapped {
			label = "d-frag-t d-frag-t-swap"
		}
		fmt.Fprintf(&b, `<text class="%s" x="51" y="%d">%s</text>`, label, barY+13, html.EscapeString(f.Name))

		if f.Swapped {
			fmt.Fprintf(&b, `<text class="d-swap-in" x="206" y="%d" text-anchor="end">swapped</text>`, barY+13)
		}
	}

	if s.WholeDoc {
		fmt.Fprintf(&b, `<text class="d-swap-tag" x="219" y="%d">whole document replaced</text>`, boxY+boxH+15)
	}

	// The exchange.
	reqY := boxY + 34
	for i, line := range s.Request {
		cls := "d-wire"
		if i > 0 {
			cls = "d-wire d-wire-hot"
		}
		fmt.Fprintf(&b, `<text class="%s" x="248" y="%d">%s</text>`, cls, reqY-32+i*15, line)
	}

	fmt.Fprintf(&b, `<line class="d-line" x1="248" y1="%d" x2="552" y2="%d" />`, reqY, reqY)
	fmt.Fprintf(&b, `<use href="#fx-arrow-r" class="d-head" x="552" y="%d" />`, reqY-4)

	respY := boxY + 84
	fmt.Fprintf(&b, `<line class="d-line d-line-dash" x1="248" y1="%d" x2="552" y2="%d" />`, respY, respY)
	fmt.Fprintf(&b, `<use href="#fx-arrow-l" class="d-head" x="248" y="%d" />`, respY-4)
	fmt.Fprintf(&b, `<text class="d-wire" x="248" y="%d">%s</text>`, respY+16, s.Response)

	// Server: the same four fragments, and what each one cost.
	fmt.Fprintf(&b, `<rect class="d-box" x="576" y="%d" width="292" height="%d" rx="9" />`, boxY, boxH)
	fmt.Fprintf(&b, `<text class="d-cap" x="588" y="%d">HANDLER</text>`, boxY+17)

	var total int
	for i, f := range s.Frags {
		barY := boxY + 26 + i*23

		cls, tcls, state := "d-frag d-frag-skip", "d-frag-t d-frag-t-skip", "skipped"
		if f.Computed {
			cls, tcls, state = "d-frag d-frag-run", "d-frag-t", "rendered"
			total += ms(f.Cost)
		}

		fmt.Fprintf(&b, `<rect class="%s" x="590" y="%d" width="182" height="18" rx="4" />`, cls, barY)
		fmt.Fprintf(&b, `<text class="%s" x="599" y="%d">%s</text>`, tcls, barY+13, html.EscapeString(f.Name))
		fmt.Fprintf(&b, `<text class="%s" x="765" y="%d" text-anchor="end">%s</text>`, tcls, barY+13, f.Cost)

		cost := "d-state"
		if !f.Computed {
			cost = "d-state d-state-skip"
		}
		fmt.Fprintf(&b, `<text class="%s" x="782" y="%d">%s</text>`, cost, barY+13, state)
	}

	fmt.Fprintf(&b, `<text class="d-total" x="765" y="%d" text-anchor="end">%s</text>`,
		boxY+boxH+16, fmt.Sprintf("%s of work", human(total)))

	// The point of the row.
	outcome := "d-out"
	if s.Fast {
		outcome = "d-out d-out-fast"
	}
	fmt.Fprintf(&b, `<text class="%s" x="896" y="%d">%s</text>`, outcome, boxY+30, html.EscapeString(s.Outcome))
	b.WriteString(wrap(s.Detail, 896, boxY+50, 29))

	b.WriteString(`</g>`)
	return b.String()
}

func ms(s string) int {
	var n int
	fmt.Sscanf(s, "%dms", &n)
	return n
}

func human(total int) string {
	if total < 1000 {
		return fmt.Sprintf("%dms", total)
	}
	return fmt.Sprintf("%d,%03dms", total/1000, total%1000)
}

// wrap lays a sentence out as <text> lines, because SVG will not do it.
func wrap(text string, x, y, width int) string {
	var b strings.Builder
	var line []string
	n := 0

	flush := func() {
		if len(line) == 0 {
			return
		}
		fmt.Fprintf(&b, `<text class="d-out-sub" x="%d" y="%d">%s</text>`, x, y, html.EscapeString(strings.Join(line, " ")))
		y += 15
		line, n = nil, 0
	}

	for _, word := range strings.Fields(text) {
		if n+len(word) > width && len(line) > 0 {
			flush()
		}
		line = append(line, word)
		n += len(word) + 1
	}
	flush()

	return b.String()
}

// HookDiagram is the first thing on the home page.
//
// It draws server-side rendering — a URL goes to a handler, HTML comes back —
// and then draws fx on top of it in the accent colour, as one extra header and
// a smaller response. That layering is the argument: the thing underneath is
// what you already have, and fx is a coat of paint on it, not a replacement.
func HookDiagram() g.Node {
	var b strings.Builder

	fmt.Fprintf(&b, `<svg class="diagram hook" viewBox="0 0 1000 376" role="img" `+
		`aria-label="A browser's URL carries the page state to a handler, which reads the query string and renders HTML. `+
		`With fx the same request adds an FX-Target header and the server answers with just that fragment.">`)

	b.WriteString(defs())
	b.WriteString(hookBrowser())
	b.WriteString(hookServer())
	b.WriteString(hookPlainLane())
	b.WriteString(hookFxLane())

	b.WriteString(`</svg>`)
	return g.Raw(b.String())
}

// hookBrowser is the address bar, labelled as what it actually is.
func hookBrowser() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<text class="k-kicker" x="0" y="14">THE BROWSER</text>`)

	// The window.
	b.WriteString(`<rect class="k-frame" x="0" y="26" width="330" height="304" rx="12" />`)
	b.WriteString(`<path class="k-chrome" d="M0 38 a12 12 0 0 1 12 -12 h306 a12 12 0 0 1 12 12 v22 h-330 z" />`)
	for i, cx := range []int{20, 34, 48} {
		fmt.Fprintf(&b, `<circle class="k-dot" cx="%d" cy="46" r="3.5" data-i="%d" />`, cx, i)
	}

	// The URL, which is the whole of the page's state.
	b.WriteString(`<rect class="k-url" x="14" y="72" width="302" height="32" rx="9" />`)
	b.WriteString(`<text class="k-urlt" x="28" y="93">/reports?status=open</text>`)
	b.WriteString(`<text class="k-callout" x="28" y="122">▲ every filter, every open pane — your page state</text>`)
	b.WriteString(`<text class="k-callout-2" x="28" y="138">and the only input your handler needs</text>`)

	// The page itself, with the one fragment fx will swap picked out.
	b.WriteString(`<rect class="k-page" x="14" y="152" width="302" height="164" rx="8" />`)
	b.WriteString(`<rect class="k-bar" x="28" y="166" width="274" height="12" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="28" y="186" width="180" height="9" rx="3" />`)

	b.WriteString(`<rect class="k-frag" x="28" y="206" width="274" height="66" rx="6" />`)
	b.WriteString(`<text class="k-fragt" x="40" y="224">#results</text>`)
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, `<rect class="k-fragbar" x="40" y="%d" width="250" height="8" rx="2" />`, 232+i*12)
	}

	b.WriteString(`<rect class="k-bar" x="28" y="284" width="210" height="9" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="28" y="301" width="150" height="9" rx="3" />`)
	b.WriteString(`</g>`)

	return b.String()
}

// hookServer is the part that was always there.
func hookServer() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<text class="k-kicker" x="700" y="14">YOUR SERVER</text>`)
	b.WriteString(`<rect class="k-frame" x="700" y="26" width="300" height="304" rx="12" />`)

	lines := []struct {
		Y    int
		Text string
		Hot  bool
	}{
		{62, "func reports(w, r) {", false},
		{84, "  f := parseFilters(r)", true},
		{106, "  rows := search(f)", false},
		{128, "  sum := rollUp(f)", false},
		{150, "  render(w, f, rows, sum)", false},
		{172, "}", false},
	}
	for _, l := range lines {
		cls := "k-code"
		if l.Hot {
			cls = "k-code k-code-hot"
		}
		fmt.Fprintf(&b, `<text class="%s" x="718" y="%d">%s</text>`, cls, l.Y, html.EscapeString(l.Text))
	}

	b.WriteString(`<line class="k-rule" x1="718" y1="196" x2="982" y2="196" />`)
	b.WriteString(`<text class="k-note" x="718" y="220">It reads the query string.</text>`)
	b.WriteString(`<text class="k-note" x="718" y="240">It renders a whole HTML page.</text>`)
	b.WriteString(`<text class="k-note k-note-dim" x="718" y="266">That is all it has ever done,</text>`)
	b.WriteString(`<text class="k-note k-note-dim" x="718" y="286">and all it has to keep doing.</text>`)
	b.WriteString(`</g>`)

	return b.String()
}

// hookPlainLane is the exchange without fx: a link, a page, a reload.
func hookPlainLane() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<text class="k-lane" x="348" y="52">WITHOUT FX — ANY BROWSER, NO JAVASCRIPT</text>`)

	b.WriteString(`<text class="k-wire" x="348" y="80">GET /reports?status=open</text>`)
	b.WriteString(`<line class="k-line" x1="348" y1="92" x2="686" y2="92" />`)
	b.WriteString(`<use href="#fx-arrow-r" class="k-head" x="686" y="88" />`)

	b.WriteString(`<line class="k-line k-line-dash" x1="348" y1="122" x2="686" y2="122" />`)
	b.WriteString(`<use href="#fx-arrow-l" class="k-head" x="348" y="118" />`)
	b.WriteString(`<text class="k-wire" x="686" y="140" text-anchor="end">200 · the whole document</text>`)

	b.WriteString(`<text class="k-outcome" x="348" y="166">The browser paints a new page. Correct, and dull.</text>`)
	b.WriteString(`</g>`)

	return b.String()
}

// hookFxLane is drawn as an overlay — same URL, same handler, one extra header
// on the way out and a smaller answer on the way back.
func hookFxLane() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<rect class="k-band" x="340" y="186" width="354" height="170" rx="12" />`)
	b.WriteString(`<text class="k-lane k-lane-fx" x="358" y="210">WITH FX — THE SAME REQUEST, PLUS ONE HEADER</text>`)

	b.WriteString(`<text class="k-wire k-wire-dim" x="358" y="238">GET /reports?status=open</text>`)
	b.WriteString(`<text class="k-wire k-wire-fx" x="358" y="254">FX-Target: #results</text>`)
	b.WriteString(`<line class="k-line k-line-fx" x1="358" y1="266" x2="676" y2="266" />`)
	b.WriteString(`<use href="#fx-arrow-r" class="k-head k-head-fx" x="676" y="262" />`)

	b.WriteString(`<line class="k-line k-line-fx k-line-dash" x1="358" y1="292" x2="676" y2="292" />`)
	b.WriteString(`<use href="#fx-arrow-l" class="k-head k-head-fx" x="358" y="288" />`)
	b.WriteString(`<text class="k-wire k-wire-fx" x="676" y="310" text-anchor="end">200 · just #results</text>`)

	// The speed claim, tied to the fragment it belongs to. The numbers are the
	// ones the lifecycle diagram further down the page works through.
	b.WriteString(`<rect class="k-chip" x="358" y="318" width="158" height="26" rx="13" />`)
	b.WriteString(`<text class="k-chipt" x="437" y="335" text-anchor="middle">1,030ms → 120ms</text>`)
	b.WriteString(`<text class="k-outcome k-outcome-fx" x="528" y="335">no reload, no flash</text>`)
	b.WriteString(`</g>`)

	return b.String()
}
