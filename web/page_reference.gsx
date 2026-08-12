package web

func ReferencePage() Page {
	return Page{
		Slug:  "reference",
		Nav:   "Reference",
		Title: "Reference",
		Lede:  "Everything fx does. It is a short page on purpose — if it were long, the library would have failed.",
		Body:  referenceBody,
	}
}

func referenceBody(c *Ctx) Node {
	return (
		<>
			{Section("attributes", "Attributes",
				Table(
					[]string{"Attribute", "On", "Effect"},
					[][]Node{
						Row(
							C("fx-target"),
							C("<a>, <form>"),
							Text("Fetch the destination, then replace these selectors with the matching elements from the response. A comma-separated list."),
						),
						Row(
							C("fx-loading-target"),
							C("<a>, <form>, <meta>"),
							Text("Add the fx-loading class to everything matching these selectors while the request is in flight, and remove it afterwards."),
						),
						Row(
							C("fx-hungry"),
							Text("any element with an id"),
							Text("Swap this element on every navigation, whether or not the link asked for it. Requires an id; without one it is ignored."),
						),
						Row(
							C("fx-interval"),
							C(`<meta name="fx-refresh">`),
							Text("Re-fetch the current URL every N milliseconds and swap fx-target."),
						),
					},
				),

				<p>
					A selector is any CSS selector, but in practice it should be an id. Two elements matching
					the same selector make the swap ambiguous, and fx picks the first one.
				</p>,
			)}

			{Section("links", "Links",
				Code("html", `
<a href="/reports?status=open" fx-target="#content">Open reports</a>
<a href="/reports" fx-target="#content, #summary" fx-loading-target="#progress-bar">All</a>
`),

				<p>
					The <code>href</code> is a real URL to a real page. fx pushes it, fetches it, and swaps.
					The address bar always shows a URL your server can render on its own.
				</p>,

				<p>fx leaves a click alone — the browser handles it normally — when:</p>,

				<ul>
					<li>it is not a left click, or Cmd, Ctrl, Shift or Alt is held</li>
					<li>the link has <code>target</code> set to anything but <code>_self</code></li>
					<li>the link has a <code>download</code> attribute</li>
				</ul>,
			)}

			{Section("forms", "Forms",
				Code("html", `
<form action="/reports" method="GET" fx-target="#results" fx-loading-target="#results">
  <input type="search" name="q">
  <button type="submit">Search</button>
</form>
`),

				<p>Form submission follows what the browser would have done:</p>,

				<ul>
					<li>
						<strong>GET</strong> puts the fields in the query string, so the URL describes the
						result — which is the point.
					</li>
					<li>
						<strong>POST</strong> sends <code>application/x-www-form-urlencoded</code>, or
						<code>multipart/form-data</code> if the form has <code>enctype="multipart/form-data"</code>
						or a file selected.
					</li>
					<li>
						The submit button's <code>name</code> and <code>value</code> are included, and only the
						button that was actually pressed.
					</li>
					<li>
						<code>formaction</code> and <code>formmethod</code> on the button override the form's.
					</li>
					<li>
						The button is disabled and given <code>fx-loading</code> for the duration of the
						request.
					</li>
					<li>
						If the response is a redirect, fx swaps the fragments from where it landed and puts
						that URL in the address bar. Post, redirect, get, without thinking about it.
					</li>
				</ul>,
			)}

			{Section("hungry", "fx-hungry",
				Code("html", `
<title id="page-title" fx-hungry>Reports — Acme</title>
<nav id="primary-nav" fx-hungry>…</nav>
`),

				<p>
					Some parts of a page change on every navigation and no link should have to know about
					them: the document title, the active state in the nav, an unread count. Mark them hungry
					and the server decides — if the response contains the element, it is swapped.
				</p>,

				<p>
					Hungry elements are collected from both the current page and the response, so a fragment
					can introduce a new hungry element and it will still be picked up.
				</p>,
			)}

			{Section("refresh", "Polling",
				Code("html", `
<meta name="fx-refresh" id="jobs-poll" fx-interval="5000" fx-target="#jobs">
`),

				<p>
					This is <code>&lt;meta http-equiv="refresh"&gt;</code> for one fragment. Every five
					seconds fx re-fetches the current URL and swaps <code>#jobs</code>.
				</p>,

				<p>Timers stop on their own, which is the part that matters:</p>,

				<ul>
					<li>when the URL changes, every timer and every in-flight request is cancelled</li>
					<li>
						when a swap removes the <code>&lt;meta&gt;</code> from the document, its timer is
						cancelled with it
					</li>
				</ul>,

				Note("Put it in the body, not the head",
					<p>
						The tag has to sit inside a fragment that gets swapped — usually inside the element
						your links target. Left in <code>&lt;head&gt;</code> it survives every navigation, and
						you will be polling a page you already left.
					</p>,
				),

				<p>
					The interval is the gap between the end of one request and the start of the next, so a
					slow server does not queue up work behind itself.
				</p>,
			)}

			{Section("loading", "Loading state",
				<p>
					<code>fx-loading-target</code> puts the <code>fx-loading</code> class on elements for the
					duration of a request. What that looks like is entirely your CSS. This site uses a bar:
				</p>,

				Code("html", `
<div id="progress-bar"></div>
<a href="/reports" fx-target="#content" fx-loading-target="#progress-bar">Reports</a>
`),

				Code("css", `
#progress-bar { height: 2px; width: 0; background: rebeccapurple; opacity: 0; }
#progress-bar.fx-loading { width: 70%; opacity: 1; transition: width 900ms; }
`),
			)}

			{Section("config", "Runtime configuration",
				<p>
					Set these on the <code>fx</code> global whenever you like. There is no init call and no
					config object.
				</p>,

				Table(
					[]string{"Property", "Default", "What it does"},
					[][]Node{
						Row(C("fx.timeout"), C("10000"), Text("Milliseconds before a request is aborted and the fallback runs.")),
						Row(C("fx.clickFallback"), C("location.replace(url)"), Text("Called when a link navigation fails.")),
						Row(C("fx.submitFallback"), C("form.submit()"), Text("Called with the form when a submission fails.")),
						Row(C("fx.historyFallback"), C("location.reload()"), Text("Called when a back or forward navigation fails.")),
						Row(C("fx.version"), C(`"`+fxVersion+`"`), Text("The version you have.")),
						Row(C("fx.logInfo / logDebug / logWarn / logError"), C("noop"), Text("Logging hooks. fx.dev.js fills them in; in production they do nothing.")),
					},
				),

				Code("js", `
fx.timeout = 30_000;
fx.clickFallback = (url, err) => {
  reportToSentry(err);
  window.location.replace(url);
};
`),
			)}

			{Section("protocol", "The server protocol",
				<p>
					Every fx request carries an <code>FX-Target</code> header listing the selectors the
					browser is about to swap:
				</p>,

				Code("http", `
GET /reports?status=open HTTP/1.1
FX-Target: #content, #primary-nav, #page-title
`),

				<p>
					You may ignore it completely. Return the whole document and everything works — that is
					the default, and it is why an existing handler needs no changes at all.
				</p>,

				<p>
					When a page has something expensive on it, read the header and skip what nobody asked
					for:
				</p>,

				Snippet("go", "handler", `
func reports(w http.ResponseWriter, r *http.Request) {
	page := Page{Filters: parseFilters(r)}

	// True on an ordinary page load, and on a fragment request that named it.
	if fx.Wants(r, "#results") {
		page.Results = search(r)        // 400ms
	}
	if fx.Wants(r, "#summary") {
		page.Summary = aggregate(r)     // 900ms
	}

	render(w, page)
}
`),

				Note("The default has to be “render it”",
					<p>
						<code>fx.Wants</code> returns true whenever the header is absent, so a page load, a
						crawler, a curl and a browser with JavaScript disabled all get the whole page. Skipping
						work is an optimisation you opt into for a request that told you it was safe.
					</p>,
				),

				<p>
					If you cache responses, vary on the header — two requests for the same URL can now
					legitimately have different bodies:
				</p>,

				Code("http", `Vary: FX-Target`),
			)}

			{Section("go", "The Go package",
				<p>
					Optional. fx is a JavaScript file and works with any server. If yours is Go, this saves
					you writing the same twenty lines:
				</p>,

				Code("bash", `go get github.com/kilianc/fragment-exchange`),

				Snippet("go", "main.go", `
import fx "github.com/kilianc/fragment-exchange"

mux.Handle("GET /fx.js", fx.Handler())     // served from the binary

fx.IsFragment(r)                            // did fx make this request?
fx.Targets(r)                               // []string{"#content", "#page-title"}
fx.Wants(r, "#results")                     // render this fragment?
fx.WantsAny(r, "#results", "#summary")      // render either?
fx.Script()                                 // the file itself, if you'd rather serve it yourself
`),
			)}

			{Section("dev", "fx.dev.js",
				<p>
					A development companion. Load it after <code>fx.js</code> and never in production:
				</p>,

				Code("html", `
<script src="/fx.js"></script>
<script src="/fx.dev.js"></script>  <!-- development only -->
`),

				<p>It turns on logging and checks every document fx parses for the silent mistakes:</p>,

				<ul>
					<li>an <code>fx-hungry</code> element with no id, which will never update</li>
					<li>duplicate ids, which make every target ambiguous</li>
					<li>
						an <code>fx-target</code> or <code>fx-loading-target</code> selector that matches
						nothing
					</li>
					<li>
						a <code>meta[name="fx-refresh"]</code> with a missing or invalid interval or target
					</li>
					<li>a <code>form[fx-target]</code> with no action</li>
				</ul>,

				<p>
					Logging is off until you ask for it — <code>fx.toggleLog(true)</code>, or
					<code>?fx-log=true</code> on any URL. The setting sticks for the tab's origin.
				</p>,
			)}

			{Section("swaps", "What a swap actually does",
				<p>Worth knowing, because it explains most surprises:</p>,

				<ul>
					<li>
						The response is parsed into a detached document. Nothing in it runs, loads or renders
						until it is put in the page.
					</li>
					<li>
						For each target, the matching element in the page is <em>replaced</em> — not its
						contents, the element. Anything the browser was holding on to inside it is gone:
						focus, selection, an open <code>&lt;details&gt;</code>, a playing video.
					</li>
					<li>
						Scroll positions inside the fragment are saved and restored by position, so a scrolled
						table survives a poll.
					</li>
					<li>
						Scripts inside a swapped fragment are re-created so they run, in order, waiting for
						external ones to load before running the next.
					</li>
					<li>
						A target that is missing from the response is skipped, and a warning appears if
						<code>fx.dev.js</code> is loaded.
					</li>
				</ul>,
			)}

			{Section("not", "What fx does not do",
				<p>
					No components, no client-side routing, no reactivity, no state management, no templates,
					no transitions, no request caching, no preloading, no partial-tree diffing, no
					WebSockets. None of these are on a roadmap, because the library's only real feature is
					that it is small.
				</p>,
			)}
		</>
	)
}
