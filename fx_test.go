package fx_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestBrowser runs fx.test.js in a real browser.
//
// fx is a few hundred lines of DOM, history and fetch. There is nothing to
// test without a browser, and nothing worth testing in a fake one — so the
// suite is a page, and this drives it: serve the repository, open the page in
// headless Chrome, and wait for it to beacon its results back.
//
// No Node, no test framework, no dependency in go.mod. Set FX_CHROME to use a
// specific browser, FX_HEADED=1 to watch it happen.
func TestBrowser(t *testing.T) {
	chrome := findChrome(t)

	results := make(chan []byte, 1)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/__fx_results", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading results: %v", err)
			return
		}
		select {
		case results <- body:
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	profile := t.TempDir()
	args := []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		// Chrome throttles timers in pages it thinks nobody is looking at,
		// which is every page in headless. The polling tests would never fire.
		"--disable-background-timer-throttling",
		"--disable-renderer-backgrounding",
		"--disable-backgrounding-occluded-windows",
		"--user-data-dir=" + profile,
		"--window-size=1200,900",
		server.URL + "/fx.test.html",
	}
	if os.Getenv("FX_HEADED") != "" {
		args = args[1:]
	}

	cmd := exec.Command(chrome, args...)
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting chrome: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	var body []byte
	select {
	case body = <-results:
	case <-time.After(90 * time.Second):
		t.Fatal("timed out waiting for the browser to report results")
	}

	var run struct {
		Passed []struct {
			Name       string `json:"name"`
			DurationMs int    `json:"durationMs"`
		} `json:"passed"`
		Failed []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
			Stack string `json:"stack"`
		} `json:"failed"`
		Skipped    []string `json:"skipped"`
		Only       bool     `json:"only"`
		DurationMs int      `json:"durationMs"`
	}
	if err := json.Unmarshal(body, &run); err != nil {
		t.Fatalf("parsing results: %v\n%s", err, body)
	}

	for _, c := range run.Passed {
		t.Run(c.Name, func(t *testing.T) {})
	}
	for _, c := range run.Failed {
		t.Run(c.Name, func(t *testing.T) {
			t.Errorf("%s\n%s", c.Error, c.Stack)
		})
	}
	for _, name := range run.Skipped {
		t.Run(name, func(t *testing.T) { t.Skip("skipped in fx.test.js") })
	}

	if len(run.Passed) == 0 && len(run.Failed) == 0 {
		t.Fatal("the browser reported no tests at all")
	}

	// An `only:` left in a commit passes CI while testing almost nothing.
	if run.Only {
		t.Error(`fx.test.js has an "only:" test, so most of the suite did not run`)
	}
}

func findChrome(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("FX_CHROME"); path != "" {
		return path
	}

	candidates := []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"}
	if runtime.GOOS == "darwin" {
		candidates = append(candidates,
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		)
	}

	for _, c := range candidates {
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				return c
			}
			continue
		}
		if path, err := exec.LookPath(c); err == nil {
			return path
		}
	}

	t.Skip("no Chrome found; set FX_CHROME to run the browser tests")
	return ""
}
