package web

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The demo is a small job runner. It has no database and no background worker:
// every job's state is a pure function of how long your session has been open,
// so it moves on its own, it is the same on every server, and it costs nothing
// to compute. What is being demonstrated is the navigation, not the jobs.

const (
	demoCookie    = "fx_demo"
	requeueCookie = "fx_requeue"
	demoCycle     = 150 * time.Second
	summaryCost   = 250 * time.Millisecond // a deliberately expensive fragment
)

type Job struct {
	ID       string
	Name     string
	Pipeline string
	Status   string // queued, running, passed, failed
	Percent  int
	Took     time.Duration
	Requeued bool
}

var jobSpecs = []struct {
	Name     string
	Pipeline string
	Seconds  int
	Offset   int
}{
	{"build-api", "api", 34, 0},
	{"unit-tests", "api", 52, 12},
	{"lint", "api", 18, 5},
	{"build-worker", "worker", 41, 27},
	{"integration-tests", "worker", 73, 40},
	{"migrate-staging", "infra", 22, 61},
	{"terraform-plan", "infra", 45, 8},
	{"image-scan", "infra", 30, 88},
	{"build-web", "web", 38, 19},
	{"visual-diff", "web", 61, 55},
	{"bundle-size", "web", 15, 103},
	{"e2e-chrome", "web", 84, 33},
	{"publish-docs", "docs", 26, 71},
	{"link-check", "docs", 12, 118},
}

// jobs computes every job's state for a session that started at start.
func jobs(start time.Time, requeued map[string]time.Time) []Job {
	now := time.Now()
	elapsed := now.Sub(start)

	out := make([]Job, 0, len(jobSpecs))
	for i, spec := range jobSpecs {
		id := strconv.Itoa(1400 + i)
		total := time.Duration(spec.Seconds) * time.Second

		job := Job{ID: id, Name: spec.Name, Pipeline: spec.Pipeline, Took: total}

		// A requeued job restarts from the moment it was requeued and then
		// settles, which is the only piece of real state in the demo.
		if at, ok := requeued[id]; ok && now.Sub(at) < total+20*time.Second {
			since := now.Sub(at)
			job.Requeued = true
			switch {
			case since < 3*time.Second:
				job.Status = "queued"
			case since < total:
				job.Status = "running"
				job.Percent = int(since * 100 / total)
			default:
				job.Status = "passed"
			}
			out = append(out, job)
			continue
		}

		phase := (elapsed + time.Duration(spec.Offset)*time.Second) % demoCycle
		cycle := int((elapsed + time.Duration(spec.Offset)*time.Second) / demoCycle)

		switch {
		case phase < 4*time.Second:
			job.Status = "queued"
		case phase < 4*time.Second+total:
			job.Status = "running"
			job.Percent = int((phase - 4*time.Second) * 100 / total)
		default:
			// Deterministic, but scattered enough to look like weather.
			if (i*7+cycle*13)%9 == 0 {
				job.Status = "failed"
			} else {
				job.Status = "passed"
			}
		}

		out = append(out, job)
	}

	return out
}

// Filters is the demo's entire state, and all of it comes from the URL.
type Filters struct {
	Query  string
	Status string
	Open   string
}

func filtersFrom(r *http.Request) Filters {
	q := r.URL.Query()
	return Filters{
		Query:  strings.TrimSpace(q.Get("q")),
		Status: q.Get("status"),
		Open:   q.Get("open"),
	}
}

// URL rebuilds the demo's address with one field changed. Every link on the
// page is built this way, which is what keeps the URL authoritative.
func (f Filters) URL(field, value string) string {
	q := url.Values{}
	set := func(k, v string) {
		if field == k {
			v = value
		}
		if v != "" {
			q.Set(k, v)
		}
	}
	set("q", f.Query)
	set("status", f.Status)
	set("open", f.Open)

	if len(q) == 0 {
		return "/demo"
	}
	return "/demo?" + q.Encode()
}

func (f Filters) match(j Job) bool {
	if f.Status != "" && f.Status != "all" && j.Status != f.Status {
		return false
	}
	if f.Query == "" {
		return true
	}
	needle := strings.ToLower(f.Query)
	return strings.Contains(strings.ToLower(j.Name), needle) ||
		strings.Contains(strings.ToLower(j.Pipeline), needle)
}

// Summary is the expensive fragment: a roll-up the poller never asks for.
type Summary struct {
	Counts map[string]int
	Total  int
}

func summarize(all []Job) Summary {
	// Stand-in for the aggregate query every dashboard has and nobody wants to
	// run four times a minute.
	time.Sleep(summaryCost)

	s := Summary{Counts: map[string]int{}, Total: len(all)}
	for _, j := range all {
		s.Counts[j.Status]++
	}
	return s
}

// --- session ------------------------------------------------------------

type session struct {
	ID    string
	Start time.Time
}

func sessionFrom(w http.ResponseWriter, r *http.Request) session {
	if c, err := r.Cookie(demoCookie); err == nil {
		if id, startUnix, ok := strings.Cut(c.Value, "."); ok {
			if unix, err := strconv.ParseInt(startUnix, 10, 64); err == nil {
				return session{ID: id, Start: time.Unix(unix, 0)}
			}
		}
	}

	now := time.Now()
	s := session{ID: strconv.FormatInt(now.UnixNano()%1e9, 36), Start: now}
	http.SetCookie(w, &http.Cookie{
		Name:     demoCookie,
		Value:    s.ID + "." + strconv.FormatInt(now.Unix(), 10),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
	return s
}

func requeuedFrom(r *http.Request) map[string]time.Time {
	out := map[string]time.Time{}

	c, err := r.Cookie(requeueCookie)
	if err != nil {
		return out
	}

	for _, part := range strings.Split(c.Value, ",") {
		id, at, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		unix, err := strconv.ParseInt(at, 10, 64)
		if err != nil {
			continue
		}
		out[id] = time.Unix(unix, 0)
	}
	return out
}

// requeueHandler is the demo's only mutation. It answers with a redirect, so
// the browser ends up on an ordinary GET of an ordinary page — and fx follows
// it and rewrites the address bar.
func requeueHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log it too, so the panel shows the whole exchange: a POST, a redirect,
	// and the GET fx follows it with.
	e := logFor(sessionFrom(w, r)).record(r, Page{Slug: "demo"})
	e.Note = "303 → redirect, fx follows it"

	id := r.FormValue("id")
	requeued := requeuedFrom(r)
	if id != "" {
		requeued[id] = time.Now()
	}

	var parts []string
	for jobID, at := range requeued {
		if time.Since(at) > 10*time.Minute {
			continue
		}
		parts = append(parts, jobID+":"+strconv.FormatInt(at.Unix(), 10))
	}
	sort.Strings(parts)

	http.SetCookie(w, &http.Cookie{
		Name:     requeueCookie,
		Value:    strings.Join(parts, ","),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	f := Filters{
		Query:  strings.TrimSpace(r.FormValue("q")),
		Status: r.FormValue("status"),
		Open:   id,
	}
	http.Redirect(w, r, f.URL("", ""), http.StatusSeeOther)
}

// --- request log --------------------------------------------------------

// logEntry is one request, as the server saw it.
type logEntry struct {
	At       time.Time
	Method   string
	Path     string
	Target   string // the raw FX-Target header, empty on a page load
	Rendered []string
	Skipped  []string
	Saved    time.Duration
	Note     string
}

func (e *logEntry) render(name string) { e.Rendered = append(e.Rendered, name) }
func (e *logEntry) skip(name string, d time.Duration) {
	e.Skipped = append(e.Skipped, name)
	e.Saved += d
}

type sessionLog struct {
	mu      sync.Mutex
	entries []*logEntry // newest first
}

const logDepth = 14

func (l *sessionLog) record(r *http.Request, p Page) *logEntry {
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}

	e := &logEntry{
		At:     time.Now(),
		Method: r.Method,
		Path:   path,
		Target: r.Header.Get("FX-Target"),
	}

	if p.Slug != "demo" {
		return e
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append([]*logEntry{e}, l.entries...)
	if len(l.entries) > logDepth {
		l.entries = l.entries[:logDepth]
	}

	return e
}

func (l *sessionLog) snapshot() []*logEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]*logEntry(nil), l.entries...)
}

// logs holds one log per browser. It is memory on a server that may be
// restarted or replaced at any moment, which is fine: losing it costs the
// visitor a few lines of history and nothing else.
var logs = struct {
	sync.Mutex
	m map[string]*sessionLog
}{m: map[string]*sessionLog{}}

const maxSessions = 2000

func logFor(s session) *sessionLog {
	logs.Lock()
	defer logs.Unlock()

	if l, ok := logs.m[s.ID]; ok {
		return l
	}

	// A demo does not deserve an eviction policy. If it fills up, start over.
	if len(logs.m) > maxSessions {
		logs.m = map[string]*sessionLog{}
	}

	l := &sessionLog{}
	logs.m[s.ID] = l
	return l
}

func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	default:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
}
