package common_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexgou/server/src/common"
)

// ── Context ────────────────────────────────────────────────────────────────────

func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := common.NewContext(w, r, nil)

	err := ctx.JSON(http.StatusOK, common.H{"key": "value"})
	if err != nil {
		t.Fatalf("JSON: unexpected error: %v", err)
	}

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body key: got %v, want %q", body["key"], "value")
	}
}

func TestContext_Param(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	ctx := common.NewContext(httptest.NewRecorder(), r, map[string]string{"id": "42"})
	if got := ctx.Param("id"); got != "42" {
		t.Errorf("Param: got %q, want %q", got, "42")
	}
	if got := ctx.Param("missing"); got != "" {
		t.Errorf("Param missing key: got %q, want empty", got)
	}
}

func TestContext_Params(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	original := map[string]string{"id": "1", "name": "alice"}
	ctx := common.NewContext(httptest.NewRecorder(), r, original)
	params := ctx.Params()

	if len(params) != 2 {
		t.Fatalf("Params len: got %d, want 2", len(params))
	}
	// Verify it's a copy — mutating it must not affect the context.
	params["id"] = "999"
	if ctx.Param("id") != "1" {
		t.Error("Params returned a reference instead of a copy")
	}
}

func TestContext_Header(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Custom", "nexgou")
	ctx := common.NewContext(httptest.NewRecorder(), r, nil)
	if got := ctx.Header("X-Custom"); got != "nexgou" {
		t.Errorf("Header: got %q, want %q", got, "nexgou")
	}
}

func TestContext_Method(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx := common.NewContext(httptest.NewRecorder(), r, nil)
	if ctx.Method() != http.MethodPost {
		t.Errorf("Method: got %q, want POST", ctx.Method())
	}
}

func TestContext_Path(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	ctx := common.NewContext(httptest.NewRecorder(), r, nil)
	if ctx.Path() != "/api/v1/users" {
		t.Errorf("Path: got %q, want /api/v1/users", ctx.Path())
	}
}

func TestContext_Body(t *testing.T) {
	payload := `{"name":"alice","age":30}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payload))
	ctx := common.NewContext(httptest.NewRecorder(), r, nil)

	var body struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := ctx.Body(&body); err != nil {
		t.Fatalf("Body: unexpected error: %v", err)
	}
	if body.Name != "alice" || body.Age != 30 {
		t.Errorf("Body: got %+v", body)
	}
}

// ── Exceptions ─────────────────────────────────────────────────────────────────

func TestHttpException_Error(t *testing.T) {
	ex := common.NewHttpException(422, "Unprocessable Entity")
	if !strings.Contains(ex.Error(), "422") {
		t.Errorf("Error() should contain status code, got: %s", ex.Error())
	}
}

func TestExceptionConstructors(t *testing.T) {
	tests := []struct {
		name   string
		ex     *common.HttpException
		status int
	}{
		{"BadRequest", common.NewBadRequestException("bad"), 400},
		{"Unauthorized", common.NewUnauthorizedException("unauth"), 401},
		{"Forbidden", common.NewForbiddenException("forbidden"), 403},
		{"NotFound", common.NewNotFoundException("not found"), 404},
		{"InternalServerError", common.NewInternalServerErrorException("oops"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.Status != tt.status {
				t.Errorf("Status: got %d, want %d", tt.ex.Status, tt.status)
			}
		})
	}
}

// ── Route builder ─────────────────────────────────────────────────────────────

type dummyGuard struct{}

func (g *dummyGuard) CanActivate(_ *common.Context) (bool, error) { return true, nil }

type dummyInterceptor struct{}

func (i *dummyInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	return next(ctx)
}

func TestRoute_Guard(t *testing.T) {
	r := common.Route{Method: "GET", Path: "/test"}
	r = r.Guard(&dummyGuard{})
	if len(r.Guards) != 1 {
		t.Errorf("Guard: got %d guards, want 1", len(r.Guards))
	}
}

func TestRoute_Intercept(t *testing.T) {
	r := common.Route{Method: "GET", Path: "/test"}
	r = r.Intercept(&dummyInterceptor{})
	if len(r.Interceptors) != 1 {
		t.Errorf("Intercept: got %d interceptors, want 1", len(r.Interceptors))
	}
}

func TestRoute_Version(t *testing.T) {
	r := common.Route{Method: "GET", Path: "/test"}
	r = r.Version("v2")
	if r.Ver() != "v2" {
		t.Errorf("Version: got %q, want v2", r.Ver())
	}
}

func TestPrintBanner_NoPanic(t *testing.T) {
	// PrintBanner writes to os.Stdout; just verify it doesn't panic.
	common.PrintBanner(common.BannerConfig{
		AppName:     "TestApp",
		Description: "Test description",
		Version:     "1.0.0",
		Environment: "test",
		Port:        "8080",
		URL:         "http://localhost:8080",
	})
}
