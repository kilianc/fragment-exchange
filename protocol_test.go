package fx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kilianc/fragment-exchange"
)

func request(target string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if target != "" {
		r.Header.Set(fx.TargetHeader, target)
	}
	return r
}

func TestTargets(t *testing.T) {
	tests := []struct {
		header string
		want   []string
	}{
		{"", nil},
		{"#content", []string{"#content"}},
		{"#content, #sidebar", []string{"#content", "#sidebar"}},
		{"  #content ,, #sidebar  ", []string{"#content", "#sidebar"}},
	}

	for _, tt := range tests {
		got := fx.Targets(request(tt.header))
		if len(got) != len(tt.want) {
			t.Fatalf("Targets(%q) = %v, want %v", tt.header, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("Targets(%q)[%d] = %q, want %q", tt.header, i, got[i], tt.want[i])
			}
		}
	}
}

func TestWants(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		selector string
		want     bool
	}{
		{"a page load wants everything", "", "#content", true},
		{"a page load wants everything, even odd selectors", "", "#anything", true},
		{"an exact match", "#content", "#content", true},
		{"a selector nobody asked for", "#content", "#sidebar", false},
		{"one of several", "#content, #sidebar", "#sidebar", true},
		{"a descendant of a requested target", "#content", "#content .row", true},
		{"not a descendant, just a prefix", "#content", "#contentious", false},

		// The client can name a descendant as its target. The container it
		// lives in still has to be rendered, or the response cannot contain it
		// and the swap finds nothing — a click that silently does nothing.
		{"the container a requested descendant lives in", "#content .row", "#content", true},
		{"a container two levels up", "#content .table .row", "#content", true},
		{"a container that only shares a prefix", "#content .row", "#contentious", false},

		// A child combinator is inside; a sibling one is beside, and rendering
		// the container does not produce it.
		{"a child combinator without spaces", "#content", "#content>.row", true},
		{"a sibling is not inside", "#content", "#content + .row", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fx.Wants(request(tt.header), tt.selector); got != tt.want {
				t.Errorf("Wants(%q, %q) = %v, want %v", tt.header, tt.selector, got, tt.want)
			}
		})
	}
}

func TestWantsAny(t *testing.T) {
	r := request("#sidebar")
	if !fx.WantsAny(r, "#content", "#sidebar") {
		t.Error("WantsAny should be true when any selector matches")
	}
	if fx.WantsAny(r, "#content", "#modal") {
		t.Error("WantsAny should be false when no selector matches")
	}
}

func TestIsFragment(t *testing.T) {
	if fx.IsFragment(request("")) {
		t.Error("a page load is not a fragment request")
	}
	if !fx.IsFragment(request("#content")) {
		t.Error("a request carrying FX-Target is a fragment request")
	}
}

func TestHandler(t *testing.T) {
	for path, want := range map[string]string{
		"/fx.js":     "window.fx",
		"/fx.dev.js": "fx.dev.js",
	} {
		w := httptest.NewRecorder()
		fx.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		if w.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, w.Code)
		}
		if got := w.Header().Get("Content-Type"); got != "text/javascript; charset=utf-8" {
			t.Errorf("%s: Content-Type = %q", path, got)
		}
		if !strings.Contains(w.Body.String(), want) {
			t.Errorf("%s: body does not look like the right file", path)
		}
	}
}
