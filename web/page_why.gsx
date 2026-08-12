package web

func WhyPage() Page {
	return Page{
		Slug:  "why",
		Nav:   "Why",
		Title: "Why fx exists",
		Lede:  "A short argument for doing much less, and an honest account of what that costs.",
		Body:  whyBody,
	}
}

func whyBody(c *Ctx) Node {
	return (
		<>
			<p>
				This started as a question I could not shake: why is building an internal tool in 2026
				harder than it was with Rails in 2008?
			</p>

			<p>
				Not slower — harder. The application I wanted was a few tables, some filters, a couple of
				forms and a detail pane. Every byte of truth in it lived in a database my server could
				reach. And the default way to build it involved a package manager, a bundler, a component
				framework, a client-side router, a data-fetching layer, a state library, and a second copy
				of my domain model written in TypeScript to keep the first one company.
			</p>

			<p>
				None of that is bad engineering. All of it exists because someone had a real problem. The
				question is whether <em>I</em> have that problem — and for a server-rendered internal tool,
				mostly, I do not.
			</p>

			{Section("cost", "The bill you are actually paying",
				<p>
					A frontend toolchain is not a one-time cost. It is a subscription, and the charges show
					up in places that do not look like frontend work:
				</p>,

				<ul>
					<li>
						<strong>Two models of the same thing.</strong> The server knows what a record is. Now
						the client knows too, slightly differently, and the bugs live in the gap.
					</li>
					<li>
						<strong>A build you have to keep alive.</strong> Toolchains rot. The version you pinned
						stops working with the Node you have, and an afternoon disappears.
					</li>
					<li>
						<strong>A dependency graph nobody has read.</strong> A modest app pulls in hundreds of
						packages, each of which can run arbitrary code on your machine at install time.
					</li>
					<li>
						<strong>A hiring and onboarding tax.</strong> My team writes Go. Asking everyone to
						also become competent in a frontend stack is expensive once, and expensive again every
						year the stack moves.
					</li>
				</ul>,

				<p>
					For a product where the client genuinely is the application — a design tool, an editor, a
					game — that bill buys something worth having. For a page that shows rows from a database
					and lets you filter them, it buys a slightly nicer transition.
				</p>,
			)}

			{Section("wanted", "What I actually wanted",
				<p>
					A website. A real one, where clicking a link asks the server for a page and the server
					sends one back. Then, on top of that and changing none of it, the ability to say
					<em>“only this part of the page is different, so only replace that part.”</em>
				</p>,

				Pull(Text("Not a client application that happens to talk to a server. A website that happens to skip the reload.")),

				<p>
					That ordering matters more than anything else on this page. If the enhanced path is the
					real path and the plain path is a degraded fallback, you have not built a website — you
					have built a client app with a courtesy exit. Sooner or later the plain path stops
					working and nobody notices, because nobody uses it.
				</p>,
			)}

			{Section("htmx", "Why not HTMX",
				<p>
					HTMX is good software and it got a lot of people to reconsider the server. It did not
					fit, and the reason is a mental model rather than a missing feature.
				</p>,

				<p>
					In HTMX, the unit is the element. An element issues a request, and the response is
					<em>that element's</em> new contents. Follow that to its conclusion and your server grows
					a route per fragment — <code>/rows</code>, <code>/row/7/status</code>,
					<code>/sidebar</code> — each returning a piece of HTML that only makes sense in the
					context of a page it never renders.
				</p>,

				Cols(
					Snippet("html", "element-centric", `
<button hx-get="/rows?page=2"
        hx-target="#rows"
        hx-swap="outerHTML">
  Next
</button>

<!-- the server needs a /rows route that
     renders rows and nothing else -->
`),
					Snippet("html", "page-centric", `
<a href="/reports?page=2"
   fx-target="#rows">
  Next
</a>

<!-- the server renders /reports, the
     same as it does for a reload -->
`),
				),

				<p>
					The practical consequence: with the element-centric model there is often no URL that
					renders the state you are looking at. You can reach page 2 by clicking, but you cannot
					link to it, refresh into it, or send it to a colleague — unless you do extra work to keep
					the address bar in sync with a state the server never renders on its own.
				</p>,

				<p>
					And the surface grows to match. <code>hx-swap</code> with a dozen strategies,
					<code>hx-swap-oob</code>, <code>hx-trigger</code> with its own event mini-language,
					<code>hx-boost</code> to make ordinary links behave like the rest, plus an extension
					system. Each piece is reasonable. Together they are a framework, and I went looking
					because I did not want one.
				</p>,
			)}

			{Section("unpoly", "Why not Unpoly",
				<p>
					Unpoly was much closer, and I want to be clear about that: it is page-centric, it treats
					real navigation as the foundation, and two of the best ideas in fx are lifted directly
					from it. Unpoly's <code>[up-hungry]</code> is <code>fx-hungry</code>. Unpoly's
					<code>X-Up-Target</code> request header is <code>FX-Target</code>. Neither is my
					invention and I would rather say so on the front page than in a footnote.
				</p>,

				<p>
					What I could not get comfortable with was how much it does for you. Layers and overlays,
					caching and preloading, form validation, transitions, compilers, a navigation model with
					opinions about history I kept having to work around. When Unpoly did what I meant, it was
					lovely. When it did not, I was reading library source to find out why my back button had
					behaved that way — and that is exactly the afternoon I was trying to stop spending.
				</p>,

				Pull(Text("A large API surface costs the same whether the code is mine or not. Eventually, managing the library is the framework work I left to avoid.")),

				<p>
					So this is a fit judgement, not a verdict. If Unpoly's defaults match how you think, use
					Unpoly — it is more capable than fx and always will be. I wanted something I could hold
					entirely in my head, and “capable” is the opposite of that.
				</p>,
			)}

			{Section("rules", "The rules fx is built on",
				<h3>1. Simplicity beats performance</h3>,
				<p>
					fx re-fetches the whole page and throws most of it away. That is wasteful and it is the
					single most important decision in the library, because it is what keeps the code at a
					size a person can read in one sitting. When you need the waste back, read the
					<code>FX-Target</code> header and render less. That choice is yours, per handler, and you
					make it when you have a reason.
				</p>,

				<h3>2. The server owns the page</h3>,
				<p>
					There is one renderer, and it is the one you already have. fx never composes, never
					templates, never decides what a page should contain. It reads a document your server
					wrote and moves elements out of it.
				</p>,

				<h3>3. The URL is the state</h3>,
				<p>
					Filters, sort order, which row is open, which modal is showing — all of it goes in the
					query string. Then a refresh reproduces the screen exactly, the back button does the
					obvious thing, and a colleague who pastes your link sees what you see. On an internal
					tool that last one is not a nicety; it is most of the job.
				</p>,

				<h3>4. Fail into the browser</h3>,
				<p>
					A timeout, a 500, a fragment missing from the response: every failure ends with the
					browser doing the navigation itself. Worst case, the user gets a page load — which is
					what they would have got anyway if you had never added the library.
				</p>,

				<h3>5. Small enough to read is small enough to change</h3>,
				<p>
					fx is one file you can open, understand and edit. It has no build step, no package, no
					lockfile and no dependency. If it does not do what you need, the fix is in front of you
					instead of in an issue tracker.
				</p>,
			)}

			{Section("supply-chain", "The part about npm",
				<p>
					There is a second reason to want a file instead of a package, and it stopped being
					theoretical some time ago. Installing a JavaScript dependency runs install scripts from
					the whole transitive graph with your user's privileges — your SSH keys, your cloud
					credentials, your browser profile. Every few months another popular package is
					compromised and everybody's CI runs the payload.
				</p>,

				<p>
					fx is a file. You can read all of it in the time it takes to read this page, copy it into
					your repository, and never think about its supply chain again, because it does not have
					one.
				</p>,
			)}

			{Section("costs", "What this costs you",
				<p>
					An honest list, because a pitch without one is marketing:
				</p>,

				<ul>
					<li>
						<strong>More bytes over the wire.</strong> Every navigation fetches a whole document to
						use part of it. On a slow connection with a heavy page, that is a real cost — until you
						implement <code>FX-Target</code>, which is the entire point of that header existing.
					</li>
					<li>
						<strong>Server round-trips for things a client could do alone.</strong> Toggling a
						disclosure or sorting a loaded table goes back to the server. Sometimes that is silly.
						Write six lines of JavaScript for those, and keep fx for navigation.
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
						lives in the browser, fx is the wrong tool and no amount of attributes will change
						that.
					</li>
				</ul>,

				Note("Where the line actually falls",
					<p>
						Every app has a corner with real local state — a spreadsheet grid, a chart, a canvas.
						Use a proper library there. fx has nothing to say about it, holds no global state, and
						will not fight you. The mistake is letting that one corner decide the architecture of
						the other twenty screens.
					</p>,
				),
			)}

			{Section("next", "Next",
				<p>
					<a href="/patterns" fx-target="#content" fx-loading-target="#progress-bar">Fragment-driven development</a>
					{" "}is how you actually build screens this way, and{" "}
					<a href="/demo" fx-target="#content" fx-loading-target="#progress-bar">the demo</a>
					{" "}is an application doing it, with a log of every request so you can watch the protocol
					work.
				</p>,
			)}
		</>
	)
}
