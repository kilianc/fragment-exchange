package web

import (
	"fmt"
	"strconv"
	"strings"
)

func DemoPage() Page {
	return Page{
		Slug:  "demo",
		Nav:   "Demo",
		Title: "A small application",
		Lede:  "A job runner, built the way this site argues for. Everything you can see is in the URL. Watch the request log while you click.",
		Wide:  true,
		Body:  demoBody,
	}
}

func demoBody(c *Ctx) Node {
	f := filtersFrom(c.R)
	all := jobs(c.start, requeuedFrom(c.R))

	// Order matters: each fragment records what it did as it is built, and the
	// log panel at the end renders that record. The page describes itself.
	summary := SummaryPanel(c, all)
	table := JobsPanel(c, f, all)
	detail := DetailPanel(c, f, all)

	return (
		<>
			{/* Inside the swapped fragment, never in <head>: navigating away
			    removes it and the timer stops with it. */}
			<meta name="fx-refresh" id="jobs-poll" fx-interval="3000" fx-target="#jobs, #detail" />

			{summary}

			<div class="demo-grid">
				<div>
					{FiltersForm(f)}
					{table}
				</div>
				<div>
					{detail}
					{LogPanel(c)}
				</div>
			</div>

			{DemoNotes()}
		</>
	)
}

// FiltersForm is a GET form, so submitting it changes the URL. That is the
// whole state management story.
func FiltersForm(f Filters) Node {
	var options []Node
	for _, s := range []string{"all", "running", "queued", "passed", "failed"} {
		opt := <option value={s}>{s}</option>
		if f.Status == s || (f.Status == "" && s == "all") {
			opt = <option value={s} selected>{s}</option>
		}
		options = append(options, opt)
	}

	return (
		<form
			class="filters"
			action="/demo"
			method="GET"
			fx-target="#jobs, #detail, #summary"
			fx-loading-target="#progress-bar"
		>
			<input type="search" name="q" value={f.Query} placeholder="filter by job or pipeline" />
			<select name="status">{Group(options)}</select>
			{If(f.Open != "", <input type="hidden" name="open" value={f.Open} />)}
			<button type="submit" class="primary">Filter</button>
			<a class="btn" href="/demo" fx-target="#jobs, #detail, #summary" fx-loading-target="#progress-bar">Reset</a>
		</form>
	)
}

// SummaryPanel is the expensive fragment. The poller never asks for it, so on
// a poll the handler never pays for it.
func SummaryPanel(c *Ctx, all []Job) Node {
	if !c.Wants("#summary") {
		c.entry.skip("#summary", summaryCost)
		return <div id="summary" class="panel" hidden></div>
	}

	c.entry.render("#summary")
	s := summarize(all)

	var cells []Node
	for _, status := range []string{"running", "queued", "passed", "failed"} {
		cells = append(cells, (
			<div>
				<span class={"pill pill-" + pillClass(status)}>{status}</span>
				<div class="mono" style="font-size:22px;margin-top:6px">{strconv.Itoa(s.Counts[status])}</div>
			</div>
		))
	}

	return (
		<div id="summary" class="panel" style="margin-bottom:20px">
			<div class="panel-head">
				roll-up
				<span class="muted" style="text-transform:none;letter-spacing:0">
					— pretend this is a 250ms aggregate query
				</span>
			</div>
			<div class="panel-body" style="display:flex;gap:36px">
				{Group(cells)}
			</div>
		</div>
	)
}

func JobsPanel(c *Ctx, f Filters, all []Job) Node {
	if !c.Wants("#jobs") {
		c.entry.skip("#jobs", 0)
		return <div id="jobs" class="panel" hidden></div>
	}
	c.entry.render("#jobs")

	var rows []Node
	for _, j := range all {
		if !f.match(j) {
			continue
		}
		rows = append(rows, JobRow(f, j))
	}

	body := <tbody>{Group(rows)}</tbody>
	if len(rows) == 0 {
		body = <tbody><tr><td colspan="4"><div class="empty">Nothing matches that filter.</div></td></tr></tbody>
	}

	return (
		<div id="jobs" class="panel">
			<div class="panel-head">jobs</div>
			<table class="jobs">
				<thead>
					<tr>
						<th>job</th>
						<th>pipeline</th>
						<th>status</th>
						<th></th>
					</tr>
				</thead>
				{body}
			</table>
		</div>
	)
}

func JobRow(f Filters, j Job) Node {
	var progress Node
	if j.Status == "running" {
		progress = <div class="bar"><span style={fmt.Sprintf("width:%d%%", j.Percent)}></span></div>
	}

	row := (
		<tr>
			{If(f.Open == j.ID, Attr("aria-selected", "true"))}
			<td class="name">
				{j.Name}
				{progress}
			</td>
			<td class="muted small">{j.Pipeline}</td>
			<td>
				<span class={"pill pill-" + pillClass(j.Status)}>{j.Status}</span>
				{If(j.Requeued, <span class="pill" style="margin-left:6px">requeued</span>)}
			</td>
			<td>
				<a
					class="small"
					href={f.URL("open", j.ID)}
					fx-target="#detail"
					fx-loading-target="#progress-bar"
				>
					open
				</a>
			</td>
		</tr>
	)

	return row
}

// DetailPanel is a pane whose open/closed state is a query parameter, so the
// back button closes it and the URL can be pasted to somebody else.
func DetailPanel(c *Ctx, f Filters, all []Job) Node {
	if !c.Wants("#detail") {
		c.entry.skip("#detail", 0)
		return <aside id="detail" class="panel" hidden></aside>
	}
	c.entry.render("#detail")

	if f.Open == "" {
		return (
			<aside id="detail" class="panel" style="margin-bottom:20px">
				<div class="panel-head">detail</div>
				<div class="panel-body">
					<p class="empty" style="padding:14px 0">Open a job to see it here.</p>
				</div>
			</aside>
		)
	}

	var job *Job
	for i := range all {
		if all[i].ID == f.Open {
			job = &all[i]
			break
		}
	}

	if job == nil {
		return (
			<aside id="detail" class="panel" style="margin-bottom:20px">
				<div class="panel-head">detail</div>
				<div class="panel-body"><p class="empty">No job {f.Open}.</p></div>
			</aside>
		)
	}

	return (
		<aside id="detail" class="panel" style="margin-bottom:20px">
			<div class="panel-head">
				job {job.ID}
				<a
					style="margin-left:auto;text-transform:none"
					href={f.URL("open", "")}
					fx-target="#detail"
					fx-loading-target="#progress-bar"
				>
					close
				</a>
			</div>
			<div class="panel-body">
				<p class="mono" style="margin-bottom:8px">{job.Name}</p>
				<p class="small muted">
					pipeline {job.Pipeline} · {job.Took.String()} ·
					<span class={"pill pill-" + pillClass(job.Status)}>{job.Status}</span>
				</p>

				{/* A POST that answers with a redirect. fx follows it, swaps
				    from where it landed, and puts that URL in the bar. */}
				<form
					action="/demo/requeue"
					method="POST"
					fx-target="#jobs, #detail, #summary"
					fx-loading-target="#progress-bar"
				>
					<input type="hidden" name="id" value={job.ID} />
					<input type="hidden" name="q" value={f.Query} />
					<input type="hidden" name="status" value={f.Status} />
					<button type="submit" class="primary">Requeue</button>
				</form>
			</div>
		</aside>
	)
}

// LogPanel is hungry: no link asks for it, and it updates on every swap.
func LogPanel(c *Ctx) Node {
	var items []Node
	for _, e := range c.log.snapshot() {
		items = append(items, LogRow(e))
	}

	if len(items) == 0 {
		items = append(items, <li class="muted">nothing yet</li>)
	}

	return (
		<div id="server-log" class="panel" fx-hungry>
			<div class="panel-head">server log</div>
			<div class="panel-body">
				<ul class="log">{Group(items)}</ul>
			</div>
		</div>
	)
}

func LogRow(e *logEntry) Node {
	target := <span class="full">page load — rendered everything</span>
	if e.Target != "" {
		target = <span>FX-Target: <span class="t">{e.Target}</span></span>
	}

	var note Node
	if e.Note != "" {
		note = <div class="t">{e.Note}</div>
	}

	var saved Node
	if len(e.Skipped) > 0 {
		saved = (
			<div class="skip">
				skipped {strings.Join(e.Skipped, ", ")}
				{If(e.Saved > 0, Text(fmt.Sprintf(" — saved ~%dms", e.Saved.Milliseconds())))}
			</div>
		)
	}

	return (
		<li>
			<div>
				<span class="m">{e.Method}</span> {truncate(e.Path, 46)}
			</div>
			<div>{target}</div>
			{note}
			{saved}
			<div class="m">{ago(e.At)}</div>
		</li>
	)
}

func DemoNotes() Node {
	return (
		<div style="max-width:760px;margin-top:48px">
			{Section("watching", "What to watch for",
				<ul>
					<li>
						<strong>The poller.</strong> Every three seconds the page re-fetches itself asking for
						<code>#jobs, #detail</code>. The log shows the roll-up being skipped each time — about
						250ms of database work this page never does.
					</li>
					<li>
						<strong>Open a job.</strong> That link asks only for <code>#detail</code>, so the table
						and the roll-up are both skipped. The URL gains <code>?open=…</code>, and the back
						button closes the pane because closing it is a URL you have already been to.
					</li>
					<li>
						<strong>Filter.</strong> A plain GET form. The fields land in the query string, the
						address bar describes what you are looking at, and a refresh reproduces it exactly.
					</li>
					<li>
						<strong>Requeue.</strong> A POST that answers <code>303 See Other</code>. fx follows the
						redirect, swaps from the page it landed on and rewrites the URL — so a refresh re-reads
						instead of re-submitting.
					</li>
					<li>
						<strong>Turn JavaScript off and use it anyway.</strong> Everything still works. It just
						reloads the page each time, which is what it was always doing underneath.
					</li>
				</ul>,

				Note("The log is the point",
					<p>
						That panel is <code>fx-hungry</code>: nothing links to it and no link asks for it, yet
						it updates on every swap. It is also the honest measure of this whole idea — the server
						is being asked for a page, every time, and it decides how much of one to build.
					</p>,
				),
			)}
		</div>
	)
}

func pillClass(status string) string {
	switch status {
	case "running":
		return "running"
	case "passed":
		return "done"
	case "failed":
		return "failed"
	default:
		return "queued"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
