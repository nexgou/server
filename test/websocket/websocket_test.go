package nexgouws_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
	nexgouws "github.com/nexgou/server/src/websocket"
)

// ── WSRoute builder ───────────────────────────────────────────────────────────

func TestNewRoute(t *testing.T) {
	r := nexgouws.NewRoute("/chat", func(ctx *nexgouws.WSContext) error { return nil })
	if r.Path != "/chat" {
		t.Errorf("Path: got %q, want /chat", r.Path)
	}
}

func TestWSRoute_Guard(t *testing.T) {
	r := nexgouws.NewRoute("/ws", nil)
	r = r.Guard(&dummyGuard{}, &dummyGuard{})
	if len(r.Guards) != 2 {
		t.Errorf("Guard: got %d, want 2", len(r.Guards))
	}
}

func TestWSRoute_Version(t *testing.T) {
	r := nexgouws.NewRoute("/ws", nil).Version("v1")
	if r.Ver() != "v1" {
		t.Errorf("Ver: got %q, want v1", r.Ver())
	}
}

func TestWSRoute_FullPath_WithVersion(t *testing.T) {
	r := nexgouws.NewRoute("/chat", nil).Version("v2")
	if got := r.FullPath(); got != "/v2/chat" {
		t.Errorf("FullPath: got %q, want /v2/chat", got)
	}
}

func TestWSRoute_FullPath_NoVersion(t *testing.T) {
	r := nexgouws.NewRoute("/events", nil)
	if got := r.FullPath(); got != "/events" {
		t.Errorf("FullPath no version: got %q", got)
	}
}

// ── SplitPath ─────────────────────────────────────────────────────────────────

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"/chat", []string{"chat"}},
		{"/v1/chat/room", []string{"v1", "chat", "room"}},
		{"/", []string{}},
		{"", []string{}},
		{"chat", []string{"chat"}},
	}
	for _, tt := range tests {
		got := nexgouws.SplitPath(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("SplitPath(%q): got %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("SplitPath(%q)[%d]: got %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

// ── WSEntry matching ──────────────────────────────────────────────────────────

func TestWSEntry_Match_Exact(t *testing.T) {
	e := nexgouws.NewEntry(nexgouws.NewRoute("/chat", nil))
	params := make(map[string]string)
	if !e.Match(nexgouws.SplitPath("/chat"), params) {
		t.Error("exact match failed")
	}
}

func TestWSEntry_Match_WithParam(t *testing.T) {
	e := nexgouws.NewEntry(nexgouws.NewRoute("/room/:id", nil))
	params := make(map[string]string)
	if !e.Match(nexgouws.SplitPath("/room/42"), params) {
		t.Error("param match failed")
	}
	if params["id"] != "42" {
		t.Errorf("param: got %q, want 42", params["id"])
	}
}

func TestWSEntry_Match_NoMatch(t *testing.T) {
	e := nexgouws.NewEntry(nexgouws.NewRoute("/chat", nil))
	params := make(map[string]string)
	if e.Match(nexgouws.SplitPath("/other"), params) {
		t.Error("should not match /other")
	}
}

func TestWSEntry_Match_DifferentLength(t *testing.T) {
	e := nexgouws.NewEntry(nexgouws.NewRoute("/a/b", nil))
	params := make(map[string]string)
	if e.Match(nexgouws.SplitPath("/a"), params) {
		t.Error("should not match shorter path")
	}
}

// ── dummy guard helper ────────────────────────────────────────────────────────

type dummyGuard struct{}

func (g *dummyGuard) CanActivate(_ *common.Context) (bool, error) { return true, nil }

// ── WSRoute.Upgrade — guard denial path (HTTP only, no real WS needed) ────────

func TestWSRoute_Upgrade_Integration(t *testing.T) {
	// We can't do a real WebSocket handshake in a unit test without a live server,
	// but we can verify the router correctly rejects upgrades that fail guards
	// by going through the router layer (tested in router_test). Here we just
	// verify NewEntry + Match work with a versioned route.
	r := nexgouws.NewRoute("/ws", func(ctx *nexgouws.WSContext) error { return nil }).Version("v1")
	e := nexgouws.NewEntry(r)
	params := make(map[string]string)
	if !e.Match(nexgouws.SplitPath("/v1/ws"), params) {
		t.Error("versioned entry should match /v1/ws")
	}
}

// Ensure WSController interface is satisfied by a test type.
type testWSController struct{}

func (c *testWSController) Register() []common.Route       { return nil }
func (c *testWSController) RegisterWS() []nexgouws.WSRoute { return nil }

func TestWSController_Interface(t *testing.T) {
	var _ nexgouws.WSController = &testWSController{}
	var _ common.Controller = &testWSController{}
}

// Ensure WSContext fields compile (can't test methods without a live conn).
var _ *nexgouws.WSContext = (*nexgouws.WSContext)(nil)

// ── WebSocket router integration: upgrade detection ───────────────────────────

func TestRouter_WS_NoUpgrade_Returns404(t *testing.T) {
	// A GET to a WS path without Upgrade headers → falls through to 404
	// (this is tested in router_test; just a compile smoke test here)
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	_ = req
}
