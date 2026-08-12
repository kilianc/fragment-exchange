package web

func IndexPage() Page {
	return Page{
		Slug:  "",
		Nav:   "Overview",
		Title: "fx — Fragment eXchange",
		Lede:  "Your server already knows how to render the page. fx swaps the parts that changed. One attribute, 400 lines, no build step.",
		Body:  indexBody,
	}
}

func indexBody(c *Ctx) Node {
	return (
		<>
			<div class="hero-cta">
				<a class="btn btn-primary" href="/demo" fx-target="#content" fx-loading-target="#progress-bar">See it running</a>
				<a class="btn" href="/why" fx-target="#content" fx-loading-target="#progress-bar">Why it exists</a>
				<a class="btn" href="/reference" fx-target="#content" fx-loading-target="#progress-bar">Reference</a>
			</div>

			{Snippet("html", "the entire library, used", `
<script src="/fx.js"></script>

<a href="/reports" fx-target="#content">Reports</a>

<main id="content">
  <!-- whatever your server rendered -->
</main>
`)}

			<p>
				Click the link and fx fetches <code>/reports</code>, pulls <code>#content</code> out of the
				response, puts it in the page and updates the address bar. No full reload, no flash, no
				scroll jump.
			</p>

			<p>
				Turn JavaScript off and the same link still works, because it is a link to a page your
				server renders. That is not a fallback bolted on afterwards. It is the starting point.
			</p>

			{Pull(Text("A site built with fx is a site that works without fx. The library only makes it quicker."))}

			{Section("what", "What it does",
				<p>Four things. That is the whole library.</p>,

				Table(
					[]string{"", "What it is for"},
					[][]Node{
						Row(C("fx-target"), Text("On a link or a form: fetch the page, swap these selectors.")),
						Row(C("fx-loading-target"), Text("Put an fx-loading class on these while the request is in flight.")),
						Row(C("fx-hungry"), Text("On any element with an id: swap me too, on every navigation.")),
						Row(C(`<meta name="fx-refresh">`), Text("Re-fetch this page on an interval and swap a fragment.")),
					},
				),

				<p>
					There is a fifth thing, and it lives on your side: every request carries an
					<code>FX-Target</code> header naming the fragments the browser is about to use. Your
					handler can read it and skip the work nobody asked for. It is entirely optional — a
					handler that ignores it is correct, just slower.
				</p>,
			)}

			{Section("how", "How a navigation works",
				Steps(
					Text("You click a link that has fx-target."),
					Text("fx pushes the URL and fetches it, with an FX-Target header."),
					Text("The server renders a page. A whole page — the same one it would render for a reload."),
					Text("fx parses the response in memory and lifts out the fragments named by fx-target."),
					Text("Those elements replace the ones in the page, and inline scripts inside them run."),
					Text("Anything goes wrong — a timeout, a 500, a missing fragment — and the browser does the navigation itself."),
				),

				<p>
					No virtual DOM, no hydration, no client-side router, no store. The page state lives in
					the URL, which means the back button, a refresh, and a link pasted into chat all show
					the same thing.
				</p>,
			)}

			{Section("install", "Install",
				<p>Copy the file. That is the install.</p>,

				Code("bash", `curl -O https://fx.ciuffolo.com/fx.js`),

				<p>
					There is no package, no lockfile and no transitive dependency to audit, because there
					is nothing to depend on. If you write Go, the repository also ships a small package
					that embeds the script and answers the header:
				</p>,

				Snippet("go", "main.go", `
mux.Handle("GET /fx.js", fx.Handler())

func reports(w http.ResponseWriter, r *http.Request) {
	var rows []Row
	if fx.Wants(r, "#results") {
		rows = expensiveQuery(r) // skipped when only the nav is being swapped
	}
	render(w, rows)
}
`),
			)}

			{Section("suits", "What it is good at, and what it is not",
				<p>
					fx is for applications whose truth lives on the server: internal tools, dashboards,
					admin panels, anything table-shaped and data-heavy. If your team writes backend code
					and your pages are already server-rendered, fx makes them feel like an application
					without making you build one.
				</p>,

				<p>
					It is not a framework and does not want to become one. There are no components, no
					reactivity, no client-side state and no template system, on purpose. When a corner of
					your page genuinely needs local state — a spreadsheet grid, a chart with a brush, a
					drag-and-drop canvas — use a real library for that corner. fx has no opinion about it
					and will not get in the way.
				</p>,

				Note("The design rule",
					<p>
						When simplicity and performance disagree, simplicity wins. fx re-fetches whole pages
						and throws most of the response away. That is wasteful, and it is exactly why the
						library fits in your head in one sitting. If you want the waste back, the
						<code>FX-Target</code> header is right there.
					</p>,
				),
			)}
		</>
	)
}
