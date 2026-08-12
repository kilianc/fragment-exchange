package web

func PatternsPage() Page {
	return Page{
		Slug:  "patterns",
		Nav:   "Patterns",
		Title: "Fragment-driven development",
		Lede:  "How to build a real application this way: the shape of the layout, where state lives, and the handful of habits that make it hold together.",
		Body:  patternsBody,
	}
}

func patternsBody(c *Ctx) Node {
	return (
		<>
			<p>
				fx is four attributes, which you can learn in a minute. What takes longer is the way of
				thinking that makes them add up to an application. It comes down to one sentence:
			</p>

			{Pull(Text("Build the whole thing as if fx did not exist. Then add fx-target."))}

			<p>
				Everything below follows from that. If a screen only makes sense after a fragment swap, it
				is not a screen — it is a state that no URL can reach, and you will pay for it the first
				time somebody refreshes.
			</p>

			{Section("layout", "One layout, one hole",
				<p>
					Start with a layout that renders a complete document, with a single element every
					navigation targets. Everything that changes per page goes inside it.
				</p>,

				Snippet("html", "layout", `
<html>
  <head>
    <title id="page-title" fx-hungry>Reports — Acme</title>
    <script src="/fx.js"></script>
  </head>
  <body>
    <div id="progress-bar"></div>

    <nav id="primary-nav" fx-hungry>
      <a href="/reports" fx-target="#content" fx-loading-target="#progress-bar">Reports</a>
      <a href="/jobs"    fx-target="#content" fx-loading-target="#progress-bar">Jobs</a>
    </nav>

    <main id="content">
      <!-- the page -->
    </main>
  </body>
</html>
`),

				<p>
					Two things are hungry: the title and the nav. Neither is inside <code>#content</code>, and
					no link mentions them — but both have to change on every navigation, and hungry is how
					you say that once instead of on every link.
				</p>,

				Note("Every page is a page",
					<p>
						Each route renders this entire document. Not a fragment, not a partial — the same
						response a browser gets on a cold load. That is the invariant the whole approach rests
						on, and it is the one that makes the back button, refresh, view-source, curl and search
						engines all work without anyone thinking about them.
					</p>,
				),
			)}

			{Section("url", "Put the state in the URL. All of it.",
				<p>
					Filters, search text, sort, page number, which row is expanded, which dialog is open.
					Everything a person can see should be reconstructible from the address bar.
				</p>,

				Snippet("html", "a filter form is just a form", `
<form action="/reports" method="GET"
      fx-target="#results"
      fx-loading-target="#results">
  <input type="search" name="q" value="{{ .Query }}">
  <select name="status">…</select>
  <button type="submit">Filter</button>
</form>

<div id="results"><!-- rows --></div>
`),

				<p>
					A GET form puts its fields in the query string, so submitting it produces{" "}
					<code>/reports?q=timeout&amp;status=open</code> in the address bar. Refresh it and you
					get the same rows. Send it to a colleague and they see what you saw. On an internal tool,
					that one property is worth more than every animation you will ever write.
				</p>,

				<p>
					It also means you have no client state to synchronise, no store to hydrate, and no bug
					where the UI and the URL disagree — because there is only one of them.
				</p>,
			)}

			{Section("detail", "Detail panes and dialogs are URL state too",
				<p>
					The instinct is to make a dialog a client-side thing: a bit of JavaScript flips a class.
					Then a refresh closes it, the back button does something surprising, and there is no way
					to link to it.
				</p>,

				<p>Make it a query parameter and all three problems disappear.</p>,

				Snippet("html", "opening a detail pane", `
<a href="/reports?q=timeout&open=1421"
   fx-target="#detail"
   fx-loading-target="#progress-bar">
  Open
</a>

<aside id="detail"><!-- server renders it open or empty --></aside>
`),

				Snippet("go", "and the handler decides", `
detail := EmptyDetail()
if id := r.URL.Query().Get("open"); id != "" {
	detail = DetailFor(id)
}
`),

				<p>
					Closing it is a link back to the same URL without <code>open</code>. The back button
					closes it for free, because closing it <em>is</em> a previous URL. For a modal, target
					both the dialog and the content behind it, and let the server render whichever
					combination the URL describes.
				</p>,
			)}

			{Section("polling", "Live data without a socket",
				<p>
					When something on the page changes on its own — a job running, a queue draining — poll
					the page you are already on and swap only the part that moves.
				</p>,

				Snippet("html", "inside #content, not in the head", `
<meta name="fx-refresh" id="jobs-poll" fx-interval="5000" fx-target="#jobs">

<div id="jobs">
  <!-- rows the server just rendered -->
</div>
`),

				<p>
					Because the tag lives inside the swapped fragment, navigating away removes it and the
					timer stops. Nothing to unsubscribe from, nothing to clean up in a lifecycle hook,
					because there is no lifecycle.
				</p>,

				<p>
					The server does not need to know it is being polled: it renders the page it always
					renders, and the client keeps the fragment it cares about. When you want to stop polling
					— the job finished — render the page without the <code>&lt;meta&gt;</code> and the timer
					disappears with it.
				</p>,
			)}

			{Section("mutations", "Mutations: post, redirect, get",
				<p>
					Write actions are forms that POST and answer with a redirect. This is the oldest pattern
					on the web and it is still the right one.
				</p>,

				Snippet("html", "a form, doing what forms do", `
<form action="/jobs/1421/requeue" method="POST"
      fx-target="#jobs, #detail"
      fx-loading-target="#progress-bar">
  <button type="submit">Requeue</button>
</form>
`),

				Snippet("go", "handler", `
func requeue(w http.ResponseWriter, r *http.Request) {
	requeueJob(r.PathValue("id"))
	http.Redirect(w, r, "/jobs?open="+r.PathValue("id"), http.StatusSeeOther)
}
`),

				<p>
					fx follows the redirect, swaps the fragments from the page it landed on, and puts that
					URL in the address bar. The result is a browser sitting on a normal GET of a normal page,
					so a refresh re-reads instead of re-submitting — and the redirect target is where you
					decide what the user should be looking at afterwards.
				</p>,

				Note("Return a page, not a message",
					<p>
						The temptation is to answer a POST with just the fragment that changed. Resist it: the
						moment a route can only produce a fragment, it is a route nobody can visit, and you
						have quietly rebuilt the thing you were avoiding.
					</p>,
				),
			)}

			{Section("cheap", "Making it cheap, once it is working",
				<p>
					Re-rendering a whole page to use one fragment is wasteful. Usually it does not matter —
					templates are fast and the expensive part is the database. When it does matter, the{" "}
					<code>FX-Target</code> header tells you exactly what you are allowed to skip.
				</p>,

				Snippet("go", "the same handler, now selective", `
func jobs(w http.ResponseWriter, r *http.Request) {
	page := JobsPage{Filters: parseFilters(r)}

	if fx.Wants(r, "#jobs") {
		page.Jobs = listJobs(r)         // 40ms
	}
	if fx.Wants(r, "#summary") {
		page.Summary = rollUp(r)        // 900ms, and it changes once an hour
	}

	render(w, page)
}
`),

				<p>
					A poll targeting <code>#jobs</code> now costs 40ms instead of 940ms, and a full page load
					still renders everything, because <code>Wants</code> is true whenever the header is
					absent. This is the one optimisation worth reaching for, and you reach for it when a
					trace tells you to — not before.
				</p>,
			)}

			{Section("javascript", "When you do need JavaScript",
				<p>
					Use it. fx is not a purity test. The rule is about scope: JavaScript for things that are
					genuinely local to the browser, the server for everything that is a fact about your data.
				</p>,

				<ul>
					<li>
						<strong>Fine:</strong> a data grid with virtual scrolling, a chart, a map, an editor,
						drag-to-reorder, a keyboard shortcut, a confirmation prompt.
					</li>
					<li>
						<strong>Not fine:</strong> keeping a copy of the rows so you can filter them without
						asking, because now there are two answers to “what is on this page” and one of them is
						stale.
					</li>
				</ul>,

				<p>
					Scripts inside a swapped fragment re-run, in order, which is usually what you want for
					initialising a widget the server just rendered. If a widget is expensive to build, keep it
					outside the fragment being swapped and let fx swap around it.
				</p>,
			)}

			{Section("mistakes", "The mistakes worth naming",
				<h3>Fragments that are not pages</h3>,
				<p>
					A route that returns half a document is a route you cannot visit, cannot link to and
					cannot debug in a browser. If you find yourself adding one, the state it represents
					probably belongs in the URL of a page you already have.
				</p>,

				<h3>Ids that are not unique</h3>,
				<p>
					Every target is resolved with <code>querySelector</code>, so a duplicate id means the swap
					hits whichever came first. <code>fx.dev.js</code> shouts about this; keep it loaded in
					development.
				</p>,

				<h3>Nesting a target inside another target</h3>,
				<p>
					If <code>#content</code> and <code>#results</code> are both swapped in one navigation and
					one contains the other, the inner swap lands on an element that is about to be discarded.
					Target one or the other.
				</p>,

				<h3>State stored in the DOM</h3>,
				<p>
					A class you toggled, a value you stashed on an element, a scroll position you tracked
					yourself: a swap replaces the element and takes all of it with them. Anything that has to
					survive a swap belongs in the URL, or outside the fragment.
				</p>,

				<h3>Targeting the whole page</h3>,
				<p>
					<code>fx-target="body"</code> works, and at that point you have written a slow reload.
					The value is in swapping the smallest thing that changed.
				</p>,
			)}

			{Section("see", "See it working",
				<p>
					<a href="/demo" fx-target="#content" fx-loading-target="#progress-bar">The demo</a> is
					every pattern on this page in one screen — a filter form, a detail pane, a poller, a POST
					that redirects — with a log of each request showing what the server was asked for and
					what it chose to render.
				</p>,
			)}
		</>
	)
}
