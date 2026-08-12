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

// HookDiagram is the first thing on the home page: what is in the browser the
// usual way, and what is in it with fx.
//
// It is deliberately not accurate about anybody's stack. It is a hook — the
// argument in one glance, before the reader has agreed to read anything. The
// honest, detailed version is LifecycleDiagram further down the page.
func HookDiagram() g.Node {
	var b strings.Builder

	fmt.Fprintf(&b, `<svg class="diagram hook" viewBox="0 0 1000 404" role="img" `+
		`aria-label="The usual way keeps a router, a store, components, hydration and a fetch client in the browser and state in two places. `+
		`With fx the browser holds one file and the state is the URL.">`)

	b.WriteString(defs())

	b.WriteString(panel(panelSpec{
		X:       0,
		Kicker:  "THE USUAL WAY",
		URL:     "/reports?status=open",
		URLNote: "",
		Blocks:  []string{"router", "store", "components", "hydration", "fetch client"},
		Wire:    "JSON",
		Server:  "API",
		Notes:   []string{"state lives in two places, and they disagree", "+ a build step, a lockfile, and a few hundred dependencies"},
	}))

	b.WriteString(panel(panelSpec{
		X:       530,
		Kicker:  "WITH FX",
		URL:     "/reports?status=open",
		URLNote: "all of your state, right here",
		Blocks:  []string{"fx.js"},
		Wire:    "HTML",
		Server:  "your server, rendering pages",
		Notes:   []string{"state lives in one place, and the server can read it", "no build step, no lockfile, nothing to install"},
		Accent:  true,
	}))

	b.WriteString(`</svg>`)
	return g.Raw(b.String())
}

type panelSpec struct {
	X       int
	Kicker  string
	URL     string
	URLNote string
	Blocks  []string
	Wire    string
	Server  string
	Notes   []string
	Accent  bool
}

func panel(p panelSpec) string {
	const w = 470

	var b strings.Builder
	x := p.X

	tone := func(base string) string {
		if p.Accent {
			return base + " " + base + "-on"
		}
		return base + " " + base + "-off"
	}

	fmt.Fprintf(&b, `<g>`)
	fmt.Fprintf(&b, `<text class="%s" x="%d" y="14">%s</text>`, tone("h-kicker"), x, p.Kicker)

	// The browser.
	fmt.Fprintf(&b, `<rect class="%s" x="%d" y="26" width="%d" height="216" rx="12" />`, tone("h-frame"), x, w)
	fmt.Fprintf(&b, `<text class="h-cap" x="%d" y="50">BROWSER</text>`, x+18)

	// The address bar, which on one side is the whole state and on the other
	// is a thing the client has to be kept in sync with.
	fmt.Fprintf(&b, `<rect class="%s" x="%d" y="62" width="%d" height="30" rx="8" />`, tone("h-url"), x+18, w-36)
	fmt.Fprintf(&b, `<text class="%s" x="%d" y="82">%s</text>`, tone("h-urlt"), x+32, html.EscapeString(p.URL))

	if p.URLNote != "" {
		fmt.Fprintf(&b, `<text class="h-urlnote" x="%d" y="82" text-anchor="end">%s</text>`, x+w-32, html.EscapeString(p.URLNote))
	}

	// What the browser is carrying.
	top := 110
	if len(p.Blocks) == 1 {
		top = 160 // one block, centred, and the empty space is the point
	}
	for i, name := range p.Blocks {
		y := top + i*24
		fmt.Fprintf(&b, `<rect class="%s" x="%d" y="%d" width="%d" height="20" rx="5" />`, tone("h-block"), x+18, y, w-36)
		fmt.Fprintf(&b, `<text class="%s" x="%d" y="%d">%s</text>`, tone("h-blockt"), x+32, y+15, html.EscapeString(name))
	}

	// The wire.
	fmt.Fprintf(&b, `<line class="%s" x1="%d" y1="242" x2="%d" y2="286" />`, tone("h-line"), x+w/2, x+w/2)
	fmt.Fprintf(&b, `<use href="#fx-arrow-r" class="%s" x="%d" y="%d" transform="rotate(90 %d 286)" />`,
		tone("h-head"), x+w/2, 282, x+w/2)
	fmt.Fprintf(&b, `<text class="%s" x="%d" y="270">%s</text>`, tone("h-wire"), x+w/2+14, p.Wire)

	// The server.
	fmt.Fprintf(&b, `<rect class="%s" x="%d" y="292" width="%d" height="46" rx="12" />`, tone("h-frame"), x, w)
	fmt.Fprintf(&b, `<text class="%s" x="%d" y="321">%s</text>`, tone("h-server"), x+18, html.EscapeString(p.Server))

	// The point.
	for i, note := range p.Notes {
		fmt.Fprintf(&b, `<text class="%s" x="%d" y="%d">%s</text>`, tone("h-note"), x, 366+i*20, html.EscapeString(note))
	}

	b.WriteString(`</g>`)
	return b.String()
}
