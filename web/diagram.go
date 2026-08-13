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

// mod adds a modifier class when the flag is set. Every diagram in the package
// draws the same shape two ways — plain, and picked out in the accent — so this
// is the shape of nearly every class attribute here.
func mod(base string, on bool, modifier string) string {
	if !on {
		return base
	}
	return base + " " + modifier
}

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
	browserClass := mod("d-box", s.WholeDoc, "d-box-hot")
	fmt.Fprintf(&b, `<rect class="%s" x="28" y="%d" width="200" height="%d" rx="9" />`, browserClass, boxY, boxH)
	fmt.Fprintf(&b, `<text class="d-cap" x="40" y="%d">BROWSER</text>`, boxY+17)

	for i, f := range s.Frags {
		barY := boxY + 26 + i*23
		cls := mod("d-frag", f.Swapped, "d-frag-swap")
		fmt.Fprintf(&b, `<rect class="%s" x="42" y="%d" width="172" height="18" rx="4" />`, cls, barY)

		label := mod("d-frag-t", f.Swapped, "d-frag-t-swap")
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
		cls := mod("d-wire", i > 0, "d-wire-hot")
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

		cost := mod("d-state", !f.Computed, "d-state-skip")
		fmt.Fprintf(&b, `<text class="%s" x="782" y="%d">%s</text>`, cost, barY+13, state)
	}

	fmt.Fprintf(&b, `<text class="d-total" x="765" y="%d" text-anchor="end">%s of work</text>`,
		boxY+boxH+16, human(total))

	// The point of the row.
	outcome := mod("d-out", s.Fast, "d-out-fast")
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

// HookDiagram is the first thing on the home page, and it has two states.
//
// It is deliberately the dumbest picture on the site, because it only has one
// thing to say: a website answers a click with a whole page, and the only
// thing fx changes is how much of the screen the browser bothers to replace.
//
// So everything that is not that is drawn identically in both states — the
// click, the server, the answer, the address, every word. The one difference
// is the blue area in the browser, which goes from the whole page to the one
// block that changed. Nothing moves, nothing is added, an area shrinks. That
// is the argument: the website did not change, the redraw did.
func HookDiagram(fxOn bool) g.Node {
	var b strings.Builder

	label := "A browser asks for a page. The server sends back a whole page, and the browser " +
		"replaces everything on the screen with it."
	if fxOn {
		label = "The same picture with fx on. The same click, the same server, the same whole page " +
			"coming back — and the browser now replaces only the one block that changed."
	}

	fmt.Fprintf(&b, `<svg class="diagram hook" viewBox="0 0 1000 380" role="img" aria-label="%s">`,
		html.EscapeString(label))

	b.WriteString(defs())

	b.WriteString(`<rect class="k-sheet" x="8" y="8" width="984" height="364" rx="16" />`)
	b.WriteString(`<text class="k-sheetlabel" x="34" y="42">HOW A WEBSITE WORKS</text>`)
	b.WriteString(`<text class="k-headline" x="34" y="74">Every website you have ever used works like this.</text>`)

	b.WriteString(hookServer(fxOn))
	b.WriteString(hookExchange())
	b.WriteString(hookBrowser(fxOn))
	b.WriteString(hookPunchline(fxOn))

	b.WriteString(`</svg>`)
	return g.Raw(b.String())
}

// hookServer is the half of the picture that fx never touches, which is why
// the only thing fx adds to it is a stamp saying so.
func hookServer(fxOn bool) string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<rect class="k-box" x="34" y="104" width="240" height="134" rx="12" />`)
	b.WriteString(`<text class="k-cap" x="52" y="130">YOUR SERVER</text>`)

	// A page, because that is the only thing it makes.
	b.WriteString(`<rect class="k-page" x="52" y="150" width="54" height="68" rx="6" />`)
	for i, w := range []int{34, 26, 34, 20} {
		fmt.Fprintf(&b, `<rect class="k-bar" x="62" y="%d" width="%d" height="6" rx="2" />`, 162+i*13, w)
	}

	b.WriteString(`<text class="k-note" x="122" y="176">Renders a whole</text>`)
	b.WriteString(`<text class="k-note" x="122" y="196">page. As always.</text>`)

	if fxOn {
		b.WriteString(`<rect class="k-chip" x="54" y="226" width="186" height="24" rx="12" />`)
		b.WriteString(`<text class="k-chipt" x="147" y="243" text-anchor="middle">nothing changed here</text>`)
	}

	b.WriteString(`</g>`)
	return b.String()
}

// hookExchange is the round trip, and it is the same round trip either way.
// The whole page comes back with fx on, exactly as it did before — that is the
// deal, and the line under it says so in both states.
func hookExchange() string {
	return `<g>
    <text class="k-wire" x="455" y="140" text-anchor="middle">you click a link</text>
    ` + arrow(630, 156, 282, 172, "k-line") + `
    ` + arrow(282, 210, 630, 196, "k-line k-line-dash") + `
    <text class="k-wire" x="455" y="236" text-anchor="middle">a whole page comes back</text>
    <text class="k-say" x="455" y="258" text-anchor="middle">every time — with fx or without it</text>
  </g>`
}

// hookBrowser holds the one thing that changes: the blue area is what the
// browser replaces. Off, it is the entire page. On, it is the list that
// changed. The page itself is drawn on top of it, identically both times.
func hookBrowser(fxOn bool) string {
	var b strings.Builder

	b.WriteString(`<g>`)
	b.WriteString(`<rect class="k-box" x="636" y="96" width="328" height="232" rx="12" />`)
	b.WriteString(`<path class="k-chrome" d="M636 108 a12 12 0 0 1 12 -12 h304 a12 12 0 0 1 12 12 v20 h-328 z" />`)
	for _, cx := range []int{654, 668, 682} {
		fmt.Fprintf(&b, `<circle class="k-dot" cx="%d" cy="115" r="3.5" />`, cx)
	}

	b.WriteString(`<rect class="k-url" x="650" y="132" width="300" height="26" rx="8" />`)
	b.WriteString(`<text class="k-urlt" x="662" y="150">/reports?status=open</text>`)

	b.WriteString(`<rect class="k-page" x="650" y="170" width="300" height="142" rx="8" />`)

	// What gets replaced.
	if fxOn {
		b.WriteString(`<rect class="k-swap" x="658" y="210" width="284" height="60" rx="6" />`)
	} else {
		b.WriteString(`<rect class="k-swap" x="656" y="176" width="288" height="130" rx="6" />`)
	}

	// The page, unchanged, over the top of it.
	b.WriteString(`<rect class="k-bar" x="666" y="184" width="150" height="11" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="666" y="203" width="96" height="7" rx="3" />`)
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, `<rect class="k-bar k-bar-list" x="666" y="%d" width="268" height="10" rx="3" />`, 220+i*15)
	}
	b.WriteString(`<rect class="k-bar" x="666" y="282" width="160" height="7" rx="3" />`)
	b.WriteString(`<rect class="k-bar" x="666" y="295" width="110" height="7" rx="3" />`)

	swapped := "all of it — replaced"
	if fxOn {
		swapped = "just this — replaced"
	}
	fmt.Fprintf(&b, `<text class="k-swapt" x="800" y="350" text-anchor="middle">%s</text>`, swapped)

	b.WriteString(`</g>`)
	return b.String()
}

// hookPunchline is what the reader is supposed to leave with, in two lines.
func hookPunchline(fxOn bool) string {
	if fxOn {
		return `<g>
    <text class="k-punch k-punch-fx" x="34" y="300">fx puts in just the part that changed.</text>
    <text class="k-punchsub" x="34" y="326">No flash. You keep your place. Your server never knew the difference.</text>
  </g>`
	}

	return `<g>
    <text class="k-punch" x="34" y="300">The browser throws it all away and paints a new one.</text>
    <text class="k-punchsub" x="34" y="326">A flash, and you lose your place — to change one list.</text>
  </g>`
}
