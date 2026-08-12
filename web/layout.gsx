package web

import "os"

// Layout renders the complete document.
//
// It always renders the complete document. fx asks for one and takes what it
// needs; a server that tried to guess which fragments to send would have to be
// right every time, and this one only has to be right once.
func Layout(c *Ctx) Node {
	title := c.Page.Title
	if c.Page.Slug != "" {
		title += " — fx"
	}

	mainClass := "content"
	if c.Page.Wide {
		mainClass += " wide"
	}

	var body Node = c.Page.Body(c)

	return (
		<>
			{Raw("<!doctype html>\n")}
			<html lang="en">
				<head>
					<meta charset="utf-8" />
					<meta name="viewport" content="width=device-width, initial-scale=1" />

					{/* fx-hungry: swapped on every navigation, whatever the link asked for */}
					<title id="page-title" fx-hungry>{title}</title>

					<meta name="description" content={c.Page.Lede} />
					<link rel="icon" href={favicon} />
					{El("style", Raw(stylesheet))}

					{/* The library this site is about, served by the package it lives in. */}
					<script src="/fx.js"></script>
					{If(!isProduction(), <script src="/fx.dev.js"></script>)}
				</head>
				<body>
					<div id="progress-bar"></div>
					<a class="skip-link" href="#content">Skip to content</a>

					<header class="top">
						<div class="top-inner">
							<a class="brand" href="/" fx-target="#content" fx-loading-target="#progress-bar">
								<span class="mark">𝑓</span>
								fragment exchange
							</a>
							{Nav(c)}
						</div>
					</header>

					<div class="shell">
						<main id="content" class={mainClass}>
							{PageHead(c)}
							{body}
							{Footer()}
						</main>
					</div>
				</body>
			</html>
		</>
	)
}

// Nav is hungry, so the current page updates even though links only ever ask
// for #content.
func Nav(c *Ctx) Node {
	var links []Node
	for _, p := range Pages() {
		if p.Nav == "" {
			continue
		}

		current := p.Slug == c.Page.Slug
		links = append(links, (
			<a
				href={p.Path()}
				fx-target="#content"
				fx-loading-target="#progress-bar"
			>
				{If(current, Attr("aria-current", "page"))}
				{p.Nav}
			</a>
		))
	}

	return (
		<nav id="primary-nav" class="primary" fx-hungry>
			{Group(links)}
			<a href="https://github.com/kilianc/fragment-exchange">GitHub</a>
		</nav>
	)
}

func PageHead(c *Ctx) Node {
	if c.Page.Title == "" {
		return nil
	}

	return (
		<div class="page-head">
			<h1>{c.Page.Title}</h1>
			{If(c.Page.Lede != "", <p class="lede">{c.Page.Lede}</p>)}
		</div>
	)
}

func Footer() Node {
	return (
		<footer class="bottom">
			<span>fx v{fxVersion} · MIT</span>
			<a href="https://github.com/kilianc/fragment-exchange">Source</a>
			<span class="muted">This page was rendered by a Go server and swapped in by the library it documents.</span>
		</footer>
	)
}

// A lightning bolt in a rounded square, as a data URI. One less request, and
// no binary in the repository.
const favicon = "data:image/svg+xml," +
	"%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E" +
	"%3Crect width='32' height='32' rx='8' fill='%234f46e5'/%3E" +
	"%3Cpath d='M18 5 L9 18 h5 l-2 9 9-13 h-5 z' fill='white'/%3E%3C/svg%3E"

func isProduction() bool {
	return os.Getenv("VERCEL_ENV") == "production" || os.Getenv("PRODUCTION") == "true"
}
