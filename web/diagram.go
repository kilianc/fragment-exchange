package web

import (
	"fmt"
	"html"
	"math"
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

	// A hungry element is part of every navigation, so it is in the header and
	// the handler has to render it. Only the fragments nobody asked for are
	// skipped.
	onlyDetail := all("#detail")
	for i := range onlyDetail {
		onlyDetail[i].Computed = onlyDetail[i].Name == "#detail" || onlyDetail[i].Name == "#primary-nav"
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
		Request:  []string{"GET /reports?status=open&amp;open=1421", "FX-Target: #detail, #primary-nav"},
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
		Request:  []string{"GET /reports?status=open&amp;open=1421", "FX-Target: #detail, #primary-nav"},
		Response: "200 · only what was asked for",
		Frags:    onlyDetail,
		Outcome:  "1,030ms → 10ms",
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
    <filter id="fx-lift" x="-20%" y="-20%" width="140%" height="140%">
      <feDropShadow dx="0" dy="6" stdDeviation="7" flood-opacity="0.13" />
    </filter>
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
// The base scene is server-side rendering and nothing else: a URL goes left to
// a handler, a whole page comes back. The two boxes are deliberately not
// aligned with each other, and the arrows are not parallel, because a picture
// on a grid reads as a specification and this one is meant to read as an idea.
//
// fx is then a card that floats above that scene and overlaps it. It changes
// nothing underneath — it adds one header on the way out and gets a smaller
// answer back, drawn in the accent colour so you can see exactly which parts
// of the picture are new.
func HookDiagram() g.Node {
	var b strings.Builder

	fmt.Fprintf(&b, `<svg class="diagram hook" viewBox="0 0 1000 512" role="img" `+
		`aria-label="A browser sends its URL and query string to a server, which renders a full page of HTML. `+
		`Layered on top, fx sends the same URL with an FX-Target header and the server renders only that fragment.">`)

	b.WriteString(defs())

	// The sheet everything already sits on.
	b.WriteString(`<rect class="k-sheet" x="18" y="64" width="964" height="430" rx="18" />`)
	b.WriteString(`<text class="k-sheetlabel" x="42" y="94">SERVER-SIDE RENDERING</text>`)
	b.WriteString(`<text class="k-equation" x="42" y="124">page = render(url)</text>`)
	b.WriteString(`<text class="k-sheetsub" x="42" y="146">what you already have</text>`)

	b.WriteString(hookServer())
	b.WriteString(hookBrowser())
	b.WriteString(hookExchange())
	b.WriteString(hookFxLayer())

	b.WriteString(`</svg>`)
	return g.Raw(b.String())
}

// hookServer sits low on the left.
func hookServer() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<rect class="k-box" x="46" y="186" width="226" height="272" rx="12" />`)
	b.WriteString(`<text class="k-cap" x="64" y="212">YOUR SERVER</text>`)

	lines := []struct {
		Y   int
		T   string
		Hot bool
	}{
		{242, "func reports(w, r) {", false},
		{264, "  f := parseFilters(r)", true},
		{286, "  rows := search(f)", false},
		{308, "  sum := rollUp(f)", false},
		{330, "  render(w, f, rows, sum)", false},
		{352, "}", false},
	}
	for _, l := range lines {
		cls := "k-code"
		if l.Hot {
			cls = "k-code k-code-hot"
		}
		fmt.Fprintf(&b, `<text class="%s" x="64" y="%d">%s</text>`, cls, l.Y, html.EscapeString(l.T))
	}

	b.WriteString(`<line class="k-rule" x1="64" y1="378" x2="254" y2="378" />`)
	b.WriteString(`<text class="k-note" x="64" y="402">It reads the query</text>`)
	b.WriteString(`<text class="k-note" x="64" y="422">string and renders</text>`)
	b.WriteString(`<text class="k-note" x="64" y="442">a page. That is all.</text>`)
	b.WriteString(`</g>`)

	return b.String()
}

// hookBrowser sits high on the right, out of step with the server on purpose.
func hookBrowser() string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<rect class="k-box" x="712" y="150" width="252" height="286" rx="12" />`)
	b.WriteString(`<path class="k-chrome" d="M712 162 a12 12 0 0 1 12 -12 h228 a12 12 0 0 1 12 12 v20 h-252 z" />`)
	for _, cx := range []int{730, 744, 758} {
		fmt.Fprintf(&b, `<circle class="k-dot" cx="%d" cy="169" r="3.5" />`, cx)
	}

	b.WriteString(`<rect class="k-url" x="726" y="196" width="224" height="30" rx="8" />`)
	b.WriteString(`<text class="k-urlt" x="738" y="216">/reports?status=open</text>`)

	b.WriteString(`<rect class="k-page" x="726" y="240" width="224" height="180" rx="8" />`)
	b.WriteString(`<rect class="k-bar" x="740" y="254" width="196" height="11" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="740" y="273" width="130" height="8" rx="3" />`)

	b.WriteString(`<rect class="k-frag" x="740" y="290" width="196" height="72" rx="6" />`)
	b.WriteString(`<text class="k-fragt" x="750" y="308">#results</text>`)
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, `<rect class="k-fragbar" x="750" y="%d" width="176" height="8" rx="2" />`, 316+i*13)
	}

	b.WriteString(`<rect class="k-bar" x="740" y="374" width="160" height="8" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="740" y="391" width="110" height="8" rx="3" />`)
	b.WriteString(`</g>`)

	return b.String()
}

// arrow draws a line with a head at the far end, at whatever angle it runs.
//
// The head is a filled shape, so it takes its own class rather than the line's
// — a dashed stroke on a solid triangle draws it hollow.
func arrow(x1, y1, x2, y2 float64, cls string) string {
	angle := math.Atan2(y2-y1, x2-x1) * 180 / math.Pi

	// The k- and p- families draw the same shapes in different diagrams.
	family := "k"
	if strings.Contains(cls, "p-line") {
		family = "p"
	}

	head := family + "-head"
	if strings.Contains(cls, "-line-fx") {
		head += " " + family + "-head-fx"
	}

	return fmt.Sprintf(
		`<line class="%s" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" />`+
			`<use href="#fx-arrow-r" class="%s" x="%.1f" y="%.1f" transform="rotate(%.2f %.0f %.0f)" />`,
		cls, x1, y1, x2, y2, head, x2-7, y2-4, angle, x2, y2)
}

// hookExchange is the round trip that was always there.
func hookExchange() string {
	var b strings.Builder

	b.WriteString(`<g>`)

	// Out: the URL, which is the whole of the input.
	b.WriteString(`<text class="k-wire" x="500" y="214" text-anchor="middle">GET /reports?status=open</text>`)
	b.WriteString(arrow(706, 230, 284, 252, "k-line"))
	b.WriteString(`<text class="k-say" x="500" y="276" text-anchor="middle">The URL and its query string are the entire input.</text>`)

	// Back: a whole page.
	b.WriteString(arrow(284, 314, 706, 298, "k-line k-line-dash"))
	b.WriteString(`<text class="k-wire" x="500" y="338" text-anchor="middle">200 — a full page of HTML</text>`)
	b.WriteString(`<text class="k-say" x="500" y="358" text-anchor="middle">The server computes the page. The browser paints it.</text>`)

	b.WriteString(`</g>`)
	return b.String()
}

// hookFxLayer is the card that floats over the scene, plus the one extra round
// trip it adds. It breaks the top of the base sheet, which is the point.
func hookFxLayer() string {
	var b strings.Builder

	b.WriteString(`<g>`)

	// The layer itself, lifted off the page.
	b.WriteString(`<g filter="url(#fx-lift)">`)
	b.WriteString(`<rect class="k-layer" x="286" y="16" width="600" height="104" rx="14" />`)
	b.WriteString(`</g>`)
	b.WriteString(`<text class="k-layerlabel" x="310" y="46">fx — A LAYER ON TOP</text>`)
	b.WriteString(`<text class="k-layertext" x="310" y="72">Change nothing underneath. Send the same URL,</text>`)
	b.WriteString(`<text class="k-layertext" x="310" y="94">and one header saying which part you will use.</text>`)
	b.WriteString(`<text class="k-header" x="310" y="112">FX-Target: #results</text>`)

	// It reaches down into the scene it is sitting on.
	b.WriteString(`<path class="k-leader" d="M770 120 C770 138 826 132 838 148" />`)

	// The second round trip, in the layer's colour.
	b.WriteString(`<text class="k-wire k-wire-fx" x="500" y="386" text-anchor="middle">GET /reports?status=open + FX-Target: #results</text>`)
	b.WriteString(arrow(706, 414, 284, 400, "k-line k-line-fx"))

	b.WriteString(arrow(284, 436, 706, 422, "k-line k-line-fx k-line-dash"))
	b.WriteString(`<text class="k-wire k-wire-fx" x="500" y="460" text-anchor="middle">200 — just #results</text>`)

	b.WriteString(`<rect class="k-chip" x="742" y="448" width="196" height="26" rx="13" />`)
	b.WriteString(`<text class="k-chipt" x="840" y="465" text-anchor="middle">same inputs, less work</text>`)
	b.WriteString(`</g>`)

	return b.String()
}
