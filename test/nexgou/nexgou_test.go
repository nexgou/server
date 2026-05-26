package nexgou_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	nexgou "github.com/nexgou/server"
)

type healthController struct{}

func newHealthController() *healthController {
	return &healthController{}
}

func (controller *healthController) Register() []nexgou.Route {
	return []nexgou.Route{
		nexgou.Get("/health", func(ctx *nexgou.Context) error {
			return ctx.JSON(http.StatusOK, nexgou.H{"status": "ok"})
		}),
	}
}

func TestPublicAPIBuildsMinimalApp(t *testing.T) {
	module := nexgou.Module(nexgou.ModuleOptions{Controllers: []any{newHealthController}})
	app := nexgou.CreateApp(module)
	app.Use(nexgou.SecurityHeaders())
	app.SetFilter(&nexgou.HttpExceptionFilter{})

	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("public middleware helper should set security headers")
	}

	body := map[string]any{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body = %+v, want status ok", body)
	}
}

func TestPublicAPIExposesHTTPHelpers(t *testing.T) {
	if nexgou.Post("/tasks", nil).Method != http.MethodPost {
		t.Fatal("Post helper should create POST route")
	}
	if nexgou.Put("/tasks/1", nil).Method != http.MethodPut {
		t.Fatal("Put helper should create PUT route")
	}
	if nexgou.Patch("/tasks/1", nil).Method != http.MethodPatch {
		t.Fatal("Patch helper should create PATCH route")
	}
	if nexgou.Delete("/tasks/1", nil).Method != http.MethodDelete {
		t.Fatal("Delete helper should create DELETE route")
	}

	exception := nexgou.BadRequestException("bad input")
	if exception.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", exception.Status)
	}

	if nexgou.ParseLoggerLevel("silent") != nexgou.LevelSilent {
		t.Fatal("ParseLoggerLevel should expose logger levels")
	}
	if nexgou.ParseLoggerFormat("json") != nexgou.FormatJSON {
		t.Fatal("ParseLoggerFormat should expose logger formats")
	}
}

func TestPublicAPIExposesLoggerAndPipes(t *testing.T) {
	log := nexgou.NewLogger(nexgou.LoggerOptions{Level: nexgou.LevelSilent, Format: nexgou.FormatJSON})
	if log.Enabled(nexgou.LevelInfo) {
		t.Fatal("silent logger should disable info logs")
	}

	value, err := (&nexgou.ParseIntPipe{}).Transform("42")
	if err != nil {
		t.Fatalf("ParseIntPipe returned error: %v", err)
	}
	if value != 42 {
		t.Fatalf("value = %v, want 42", value)
	}
}

func TestPublicListenAndServeCompiles(t *testing.T) {
	var _ func(string, *nexgou.App) error = nexgou.ListenAndServe
}

func TestPublicAPIExposesStartupPrinters(t *testing.T) {
	module := nexgou.Module(nexgou.ModuleOptions{Controllers: []any{newHealthController}})
	app := nexgou.CreateApp(module)

	var buffer bytes.Buffer
	app.WriteRoutes(&buffer)
	if !strings.Contains(buffer.String(), "Mapped routes:") || !strings.Contains(buffer.String(), "/health") {
		t.Fatalf("routes output = %s, want mapped /health route", buffer.String())
	}

	nexgou.PrintRoutes(app)
	nexgou.PrintBanner(nexgou.BannerConfig{AppName: "TestApp", Version: "2.0.0"})
}
