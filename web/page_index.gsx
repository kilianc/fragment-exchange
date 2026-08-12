package web

func IndexPage() Page {
	return Page{
		Slug:    "",
		Nav:     "Overview",
		Title:   "fx — Fragment eXchange",
		Tagline: "Server-rendered web apps like it's 2005, with navigation from this decade.",
		Lede:    "For backend engineers who want an application, not a frontend project. No build step, no client state, no framework — every filter, every open pane and every sort order lives in the URL, your handler reads it and renders HTML the way it always has, and fx swaps the fragment that changed. One attribute, about 400 lines, nothing to install.",
		Body:    indexBody,
	}
}

func indexBody(c *Ctx) Node {
	return (
		<>
			<div class="figure-wide hero-figure">
				<div class="diagram-scroll">
					{HookDiagram()}
				</div>
				<p class="figure-caption">
					A page is a render function, and the URL is its only argument — which is precisely why
					everything a person can see has to live in there. fx does not change that function. It
					adds one header to the call, so the function can skip the parts of its own answer that
					nobody is going to look at.
				</p>
			</div>

			<div class="hero-cta">
				<a class="btn btn-primary" href="/demo" fx-target="#content" fx-loading-target="#progress-bar">See it running</a>
				<a class="btn" href="/reference" fx-target="#content" fx-loading-target="#progress-bar">Reference</a>
				<a class="btn" href="https://github.com/kilianc/fragment-exchange">GitHub</a>
			</div>

			{Snippet("html", "the entire library, used", `
<script src="/fx.js"></script>

<a href="/reports" fx-target="#content">Reports</a>

<main id="content">
  <!-- whatever your server rendered -->
</main>
`)}

			<p>
				Click the link. fx fetches <code>/reports</code>, lifts <code>#content</code> out of the
				response, puts it in the page and updates the address bar. No reload, no flash, no scroll
				jump.
			</p>

			<p>
				Turn JavaScript off. The same link still works, because it is a link to a page your server
				renders.
			</p>

			<p>
				That is the entire library. Everything below is the argument for why it should never get
				any bigger.
			</p>

			{Pull(Text("State lives in two places: the URL, and your database. There is no third place for a bug to hide."))}

			{Section("baseline", "The baseline is already good",
				<p>
					Server-side rendering is simple, observable and robust. It works with your cache, your
					load balancer, your logs and the application logic you already wrote. Every engineer on
					your team understands it without being taught. A full page reload is correct, trivial to
					reason about, and impossible to get subtly wrong.
				</p>,

				<p>
					It has exactly one problem: reloading the whole document to change 5% of it feels bad. A
					flash, a lost scroll position, a re-run of everything the page needed the first time.
				</p>,

				<p>
					That is a small problem. The industry solved it with a very large answer.
				</p>,
			)}

			{Section("cost", "The cost of accidental complexity",
				<p>
					Modern web development assumes a client-side framework is the default, even for
					applications whose data and operational requirements are inherently server-centric. Teams
					carry the burden of a frontend toolchain, a dependency ecosystem and a dual-state
					architecture — while getting none of the leverage those systems were designed to provide.
				</p>,

				<p>
					It is not a one-time cost. It is a subscription, and the charges land in places that do
					not look like frontend work:
				</p>,

				<ul>
					<li>
						<strong>Two models of the same thing.</strong> Your server knows what a record is. Now
						the client knows too, slightly differently, and the bugs live in the gap.
					</li>
					<li>
						<strong>A build you have to keep alive.</strong> Toolchains rot. The version you pinned
						stops working with the runtime you have, and an afternoon disappears.
					</li>
					<li>
						<strong>A dependency graph nobody has read.</strong> A modest app pulls in hundreds of
						packages, any of which can run arbitrary code on your laptop at install time.
					</li>
					<li>
						<strong>An onboarding tax, forever.</strong> Teaching a backend team a second stack is
						expensive once, and expensive again every year that stack moves.
					</li>
				</ul>,

				<p>
					For a product where the client genuinely <em>is</em> the application — an editor, a design
					tool, a game — that bill buys something worth having. For a page that shows rows from a
					database and lets you filter them, it buys a slightly nicer transition.
				</p>,
			)}

			{browserFreebies()}

			{Section("wanted", "What I actually wanted",
				<p>
					The developer experience of Rails, or PHP, or anything from the era when the server did
					the work and a page was a page. My team writes Go. They are excellent at backends and
					have no interest in becoming frontend engineers, and I did not think they should have to
					be in order to ship an internal tool.
				</p>,

				<p>
					So the starting point of a site built with fx is that it is <em>not</em> built with fx.
					It is a fully working, JavaScript-free website. You click, the server re-renders, the
					page reloads. Then, on top of that and changing none of it, a handful of attributes let
					the browser skip the reload and replace only what moved.
				</p>,

				Pull(Text("A site built with fx is a site that works without fx. The library only makes it quicker.")),

				<p>
					That ordering is the whole design. If the enhanced path is the real path and the plain
					path is a courtesy fallback, you have not built a website — you have built a client app
					with an emergency exit that nobody tests. Here the plain path is the only path. fx is a
					shortcut across it.
				</p>,
			)}

			{Section("htmx", "Why not HTMX",
				<p>
					HTMX is good software, and it got a lot of people to reconsider the server. It did not
					fit, and the reason is a mental model rather than a missing feature.
				</p>,

				<p>
					In HTMX the unit is the element. An element issues a request and receives its own
					replacement. Follow that to its conclusion and your server grows a route per fragment —{" "}
					<code>/rows</code>, <code>/row/7/status</code>, <code>/sidebar</code> — each returning
					HTML that only makes sense inside a page it never renders.
				</p>,

				Cols(
					Snippet("html", "element-centric", `
<button hx-get="/rows?page=2"
        hx-target="#rows"
        hx-swap="outerHTML">
  Next
</button>

<!-- needs a /rows route that renders
     rows and nothing else -->
`),
					Snippet("html", "page-centric", `
<a href="/reports?page=2"
   fx-target="#rows">
  Next
</a>

<!-- renders /reports, exactly as it
     does for a reload -->
`),
				),

				<p>
					The consequence is not stylistic. With the element-centric model there is often no URL
					that renders what you are looking at. You can reach page 2 by clicking, but you cannot
					link to it, refresh into it, or send it to a colleague — not without extra work to keep
					the address bar in sync with a state the server cannot produce on its own.
				</p>,

				<p>
					And the surface grows to match. <code>hx-swap</code> with a dozen strategies,{" "}
					<code>hx-swap-oob</code>, <code>hx-trigger</code> with its own event mini-language,{" "}
					<code>hx-boost</code> to make ordinary links behave like everything else, an extension
					system. Every piece is reasonable. Together they are a framework, and I went looking
					precisely because I did not want one.
				</p>,
			)}

			{Section("unpoly", "Why not Unpoly",
				<p>
					Unpoly was much closer, and credit where it is due: it is page-centric, it treats real
					navigation as the foundation, and two of the best ideas in fx are lifted straight from
					it. Unpoly's <code>[up-hungry]</code> is <code>fx-hungry</code>. Unpoly's{" "}
					<code>X-Up-Target</code> header is <code>FX-Target</code>. Neither is my invention, and I
					would rather say so on the front page than in a footnote.
				</p>,

				<p>
					What I could not get comfortable with was how much it does for you. Layers and overlays,
					caching and preloading, form validation, transitions, compilers, and a navigation model
					with firm opinions about history that I kept working around. When Unpoly did what I
					meant, it was lovely. When it did not, I was reading library source to find out why my
					back button had behaved that way — which is the exact afternoon I was trying to stop
					spending.
				</p>,

				Pull(Text("A wide API costs the same whether the code is mine or not. Past a certain point, managing the library is the framework work I left to avoid.")),

				<p>
					This is a fit judgement, not a verdict. If Unpoly's defaults match how you think, use
					Unpoly — it is more capable than fx and always will be. I wanted something I could hold
					entirely in my head, and “capable” is the opposite of that.
				</p>,
			)}

			{Section("lifecycle", "The lifecycle",
				<p>
					One address, one handler, three levels of adoption. Each row is a place you are allowed
					to stop.
				</p>,

				<div class="figure-wide">
					<div class="diagram-scroll">
						{LifecycleDiagram()}
					</div>
					<p class="figure-caption">
						Level 1 is the row people do not believe until they see it written down: the handler is
						untouched, renders exactly what it always rendered, and you still get the navigation.
						Level 2 is the optimisation you reach for once a fragment has grown expensive enough to
						be a component of its own — a modal, a widget, a roll-up nobody is looking at.
					</p>
				</div>,

				<p>
					That is what makes it incremental. The backend keeps doing the thing backends are good at
					— rendering a complete page for a URL — and the client gets faster in two steps, neither
					of which is required, and neither of which can leave you with a screen that has no
					address.
				</p>,
			)}

			{Section("rules", "The rules",
				<h3>1. Simplicity beats performance. Always.</h3>,
				<p>
					fx re-fetches the whole page and throws most of it away. That is wasteful, and it is the
					most important decision in the library, because it is what keeps the code small enough to
					read in one sitting. Want the waste back? Read the header and render less. Your call,
					per handler, when you have a reason.
				</p>,

				<h3>2. The server owns the page</h3>,
				<p>
					One renderer, and it is the one you already have. fx never composes, never templates,
					never decides what a page should contain. It reads a document your server wrote and
					moves elements out of it. The browser is a display layer.
				</p>,

				<h3>3. The URL is king</h3>,
				<p>
					Filters, sort, page, which row is open, which dialog is showing — all of it in the query
					string. Then a refresh reproduces the screen exactly, the back button does the obvious
					thing, and a colleague who pastes your link sees what you see. On an internal tool that
					last one is not a nicety. It is most of the job.
				</p>,

				<h3>4. Fail into the browser</h3>,
				<p>
					A timeout, a 500, a fragment missing from the response — every failure ends with the
					browser doing the navigation itself. Worst case the user gets a page load, which is what
					they would have got if you had never added the library.
				</p>,

				<h3>5. Small enough to read is small enough to change</h3>,
				<p>
					One file. No build step, no package, no lockfile, no dependency. If fx does not do what
					you need, the fix is in front of you rather than in somebody's issue tracker. The goal is
					a library so dumb that thirty seconds of documentation is enough.
				</p>,
			)}

			{Section("api", "The whole API",
				<p>Four attributes. That is not an excerpt.</p>,

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
					And one thing on your side. Every request carries an <code>FX-Target</code> header naming
					the fragments the browser is about to use, so your handler can skip the work nobody asked
					for. It is optional. A handler that ignores it is correct, just slower.
				</p>,

				Snippet("go", "handler", `
func reports(w http.ResponseWriter, r *http.Request) {
	page := ReportsPage{Filters: parseFilters(r)}

	// True on any page load, and on a fragment request that named it.
	if fx.Wants(r, "#summary") {
		page.Summary = aggregate(r) // 900ms, skipped when nobody asked
	}

	render(w, page)
}
`),

				<p>
					<a href="/reference" fx-target="#content" fx-loading-target="#progress-bar">The reference</a>
					{" "}is one page, and it is short for the same reason.
				</p>,
			)}

			{Section("costs", "What this costs you",
				<p>An honest list, because a pitch without one is marketing:</p>,

				<ul>
					<li>
						<strong>More bytes over the wire.</strong> Every navigation fetches a whole document to
						use part of it. On a slow connection with a heavy page that is a real cost — until you
						answer the header, which is exactly why the header exists.
					</li>
					<li>
						<strong>Round-trips for things a client could do alone.</strong> Toggling a disclosure
						or sorting a table you already have is a server request. Sometimes that is silly. Write
						six lines of JavaScript for those and keep fx for navigation.
					</li>
					<li>
						<strong>No component model.</strong> Reuse is whatever your server-side templating
						gives you. If your templating is bad, fx will not save you.
					</li>
					<li>
						<strong>Element identity is not preserved.</strong> A swap replaces elements outright,
						so a focused input or an open <code>&lt;details&gt;</code> inside a swapped fragment is
						gone. fx restores scroll positions and nothing else. Target more narrowly, or keep the
						stateful part outside the fragment.
					</li>
					<li>
						<strong>It will not carry a rich client.</strong> If your application's state genuinely
						lives in the browser, fx is the wrong tool and no amount of attributes will fix that.
					</li>
				</ul>,

				Note("Where the line actually falls",
					<p>
						Every app has a corner with real local state — a spreadsheet-like grid over ten thousand
						rows, a chart with a brush, a canvas. Use a proper library there. Use React there if you
						want. fx holds no global state, claims no ownership of the DOM, and will not fight you.
						The mistake is letting that one corner decide the architecture of the other twenty
						screens.
					</p>,
				),
			)}

			{Section("npm", "The part about npm",
				<p>
					There is a second reason to want a file instead of a package, and it stopped being
					theoretical a while ago. Installing a JavaScript dependency runs install scripts from the
					entire transitive graph with your user's privileges — your SSH keys, your cloud
					credentials, your browser profile. Every few months another popular package is
					compromised and everybody's CI runs the payload.
				</p>,

				<p>
					fx is a file. Read all of it in less time than it took to read this page, copy it into
					your repository, and never think about its supply chain again, because it does not have
					one.
				</p>,
			)}

			{Section("install", "Install",
				<p>Copy the file. That is the install.</p>,

				Code("bash", `curl -O https://fx.ciuffolo.com/fx.js`),

				<p>
					If you write Go, the repository also ships a small package that embeds the script and
					answers the header:
				</p>,

				Code("bash", `go get github.com/kilianc/fragment-exchange`),

				Snippet("go", "main.go", `
import fx "github.com/kilianc/fragment-exchange"

mux.Handle("GET /fx.js", fx.Handler())  // served from your binary

fx.IsFragment(r)                        // did fx make this request?
fx.Targets(r)                           // []string{"#content", "#page-title"}
fx.Wants(r, "#summary")                 // render this fragment?
`),
			)}

			{Section("next", "Where to go next",
				<ul>
					<li>
						<a href="/patterns" fx-target="#content" fx-loading-target="#progress-bar">Fragment-driven development</a>
						{" "}— the shape of a real application built this way: where state lives, how dialogs and
						detail panes work, and the mistakes worth naming.
					</li>
					<li>
						<a href="/reference" fx-target="#content" fx-loading-target="#progress-bar">Reference</a>
						{" "}— everything fx does, on one short page.
					</li>
					<li>
						<a href="/demo" fx-target="#content" fx-loading-target="#progress-bar">The demo</a>
						{" "}— a working application with a log of every request, showing what the browser asked
						for against what the server chose to render.
					</li>
				</ul>,

				<p class="muted small">
					This page was rendered by a Go server and swapped in by the library it describes. If you
					got here by clicking, you have already seen it work.
				</p>,
			)}
		</>
	)
}

// browserFreebies is the inventory: everything a document gets for nothing,
// and therefore everything you take ownership of the moment you stop shipping
// documents. It is deliberately exhaustive. The length is the argument.
func browserFreebies() Node {
	return Section("free", "What the browser already does for you",
		<p>
			The bill on the previous section is abstract. Here is the concrete version — the things a
			browser does for a plain HTML page, for free, that a client-side application has to
			reimplement, approximate, or quietly do without.
		</p>,

		<h3>Navigation and history</h3>,
		<ul>
			<li><strong>Back and forward.</strong> Correct by construction, because each screen was a real navigation.</li>
			<li><strong>Scroll restoration.</strong> Going back returns you to where you were on the page, not to the top of it.</li>
			<li><strong>Instant back and forward.</strong> The bfcache can restore a whole previous page without a request.</li>
			<li><strong>Bookmarks.</strong> A screen you can save and come back to next month.</li>
			<li><strong>Open in a new tab.</strong> Cmd-click, middle-click, right-click, drag to the tab bar.</li>
			<li><strong>Session restore.</strong> Reopen the browser after a crash and your tabs come back, on the right screens.</li>
			<li><strong>History search.</strong> Type half a remembered word into the address bar and find the page again.</li>
			<li><strong>Refresh.</strong> A guarantee that what you are looking at can be produced again from scratch.</li>
		</ul>,

		<h3>Addressability</h3>,
		<ul>
			<li><strong>A URL for every screen.</strong> Which is to say: a way to talk about a screen at all.</li>
			<li><strong>Query strings.</strong> Search, filters, sort order and pagination, with a format everything understands.</li>
			<li><strong>Fragment anchors.</strong> Deep links into a section of a long page.</li>
			<li><strong>The document title.</strong> Names your tab, your history entry, your bookmark and your window switcher.</li>
			<li><strong>Link previews.</strong> Paste the URL into Slack and something sensible unfurls.</li>
		</ul>,

		<h3>Sending data to the server</h3>,
		<ul>
			<li><strong>Forms.</strong> GET puts the fields in the query string, POST puts them in the body. No request code.</li>
			<li><strong>The right encoding.</strong> Urlencoded, or multipart when there is a file. Chosen for you, correctly.</li>
			<li><strong>File upload.</strong> The operating system's file picker, streamed to the server, with progress.</li>
			<li><strong>Native validation.</strong> <code>required</code>, <code>type="email"</code>, <code>pattern</code>, <code>min</code> and <code>max</code> — with messages already translated into the user's language.</li>
			<li><strong>Enter to submit.</strong> And the focus behaviour that goes with it.</li>
			<li><strong>Knowing which button was pressed.</strong> The submitter's name and value, and only that submitter's.</li>
			<li><strong>Password managers.</strong> They find the fields, fill them, and offer to save what you typed.</li>
			<li><strong>Autofill.</strong> Addresses, cards, and one-time codes lifted out of an incoming SMS.</li>
		</ul>,

		<h3>Loading, and what happens when it fails</h3>,
		<ul>
			<li><strong>The loading indicator.</strong> The spinner in the tab, and a stop button that works.</li>
			<li><strong>Streaming.</strong> HTML paints as it arrives. Nothing waits for the last byte.</li>
			<li><strong>Network error pages.</strong> Written, tested and translated by people who are not you.</li>
			<li><strong>Offline.</strong> A state the browser understands and tells the user about.</li>
			<li><strong>Transport-level retries and timeouts.</strong> Under the request, where they belong.</li>
		</ul>,

		<h3>The network</h3>,
		<ul>
			<li><strong>HTTP caching.</strong> <code>Cache-Control</code>, <code>ETag</code> and conditional requests — a caching layer you configure rather than write.</li>
			<li><strong>Compression.</strong> Negotiated per request, applied to your HTML.</li>
			<li><strong>Connection reuse and multiplexing.</strong> Free with HTTP/2, and nothing to tune.</li>
			<li><strong>Redirects.</strong> Followed, with the address bar updated to where you actually landed.</li>
			<li><strong>Cookies.</strong> Scoping, expiry, and <code>HttpOnly</code>, <code>SameSite</code> and <code>Secure</code> meaning what they say.</li>
			<li><strong>Authentication challenges.</strong> Handled by the browser, including the credential storage.</li>
			<li><strong>Range requests.</strong> Resumable downloads and seekable media.</li>
			<li><strong>Preload and prefetch hints.</strong> Declared in markup, scheduled by the browser.</li>
		</ul>,

		<h3>Input and accessibility</h3>,
		<ul>
			<li><strong>Focus order.</strong> Tab moves through the document in source order, without a manager.</li>
			<li><strong>Keyboard activation.</strong> Enter on a link, Space on a button, Escape on a dialog.</li>
			<li><strong>Landmarks and roles.</strong> Free with real elements, and hand-written once everything is a <code>&lt;div&gt;</code>.</li>
			<li><strong>Page-change announcements.</strong> A navigation tells assistive technology that something happened.</li>
			<li><strong>Find in page.</strong> Cmd-F searches text that is actually in the document.</li>
			<li><strong>Selection and copy.</strong> Including copying a table into a spreadsheet and having it paste as a table.</li>
			<li><strong>Zoom and reflow.</strong> Text scales and the layout keeps working.</li>
			<li><strong>User preferences.</strong> Reduced motion, high contrast, forced colours, minimum font size.</li>
			<li><strong>The context menu.</strong> Copy link address, open in new tab, search for this selection.</li>
		</ul>,

		<h3>The life of a page after you ship it</h3>,
		<ul>
			<li><strong>Search engines.</strong> They read documents. Reliably, and without a rendering budget.</li>
			<li><strong>Reader mode.</strong> Which needs real article markup to find anything.</li>
			<li><strong>Translation.</strong> Browser and extension translators walk the DOM of a document.</li>
			<li><strong>Print.</strong> Including a print stylesheet, which is a few lines of CSS on a real page.</li>
			<li><strong>View source.</strong> The oldest debugging tool on the web, and still one of the best.</li>
			<li><strong>curl, wget and scripts.</strong> Anything that can make a request can use your app.</li>
			<li><strong>Archives.</strong> A page that a crawler can save is a page that outlives your hosting.</li>
		</ul>,

		<h3>Security defaults</h3>,
		<ul>
			<li><strong>Same-origin policy.</strong> Enforced on a document you did not have to configure.</li>
			<li><strong>Content Security Policy.</strong> A header, applied to the whole page.</li>
			<li><strong>Referrer policy.</strong> Declared once, obeyed everywhere.</li>
			<li><strong>Mixed content blocking.</strong> Automatic, and loud.</li>
			<li><strong>Subresource integrity.</strong> An attribute, not a build step.</li>
		</ul>,

		<p>
			That is fifty-five, and I am certain I have missed some.
		</p>,

		Pull(Text("None of these are impossible to rebuild. That is not the claim. The claim is that you now own all of them, forever, and you will get some of them wrong.")),

		<p>
			A good client-side framework will hand you a router that fixes the back button and a library
			that manages focus. What it cannot hand you is the other fifty — and the ones it misses fail
			quietly, in the hands of the user who needed them most: the one on a screen reader, the one
			pasting a link to a colleague, the one whose connection dropped halfway through.
		</p>,

		<p>
			So the motivation behind fx is not nostalgia, and it is not minimalism for its own sake. It is
			that this list is enormous, the platform implements all of it already, and the price of
			keeping every item on it is to keep shipping documents at URLs. That is the whole bargain.
			fx exists so you can take the one thing a reload costs you — the flash — without paying for
			it with the list.
		</p>,

		Note("The one thing a swap does interrupt",
			<p>
				A fragment swap is not a navigation, so nothing announces it. A screen reader is told when
				a page loads and told nothing when an element is replaced. If a swap changes something a
				user needs to know about, say so — an <code>aria-live</code> region on the fragment, or
				moving focus to it. It is the one item on the list fx hands back to you, and it is worth
				knowing that it is exactly one.
			</p>,
		),
	)
}
