package benchmark_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	benchapp "github.com/nexgou/server/benchmark/nexgou/app"
)

func TestNexGouBenchmarkHealth(t *testing.T) {
	handler, closeStore := newBenchmarkHandler(t)
	defer closeStore()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}

	body := map[string]any{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body["status"] != "ok" || body["service"] != "nexgou" || body["version"] != "2.0.0" {
		t.Fatalf("health body = %+v", body)
	}
}

func TestNexGouBenchmarkCRUDContract(t *testing.T) {
	handler, closeStore := newBenchmarkHandler(t)
	defer closeStore()

	created := requestJSON(t, handler, http.MethodPost, "/users", map[string]any{
		"name":  "Sergio Gonzalez",
		"email": "sergio@example.com",
		"age":   34,
	}, http.StatusCreated)
	id := int(created["id"].(float64))

	fetched := requestJSON(t, handler, http.MethodGet, "/users/"+itoa(id), nil, http.StatusOK)
	if fetched["email"] != "sergio@example.com" {
		t.Fatalf("fetched user = %+v", fetched)
	}

	listed := requestJSON(t, handler, http.MethodGet, "/users?limit=20&offset=0", nil, http.StatusOK)
	if listed["total"] != float64(1) {
		t.Fatalf("list response = %+v, want total 1", listed)
	}

	updated := requestJSON(t, handler, http.MethodPut, "/users/"+itoa(id), map[string]any{
		"name":  "Sergio Updated",
		"email": "sergio.updated@example.com",
		"age":   35,
	}, http.StatusOK)
	if updated["age"] != float64(35) {
		t.Fatalf("updated user = %+v", updated)
	}

	deleted := requestJSON(t, handler, http.MethodDelete, "/users/"+itoa(id), nil, http.StatusOK)
	if deleted["deleted"] != true {
		t.Fatalf("delete response = %+v", deleted)
	}

	requestJSON(t, handler, http.MethodGet, "/users/"+itoa(id), nil, http.StatusNotFound)
}

func newBenchmarkHandler(t *testing.T) (http.Handler, func()) {
	t.Helper()
	store, err := benchapp.NewStore(filepath.Join(t.TempDir(), "db.sqlite"))
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	config := benchapp.Config{ServiceName: "nexgou", Version: "2.0.0"}
	app := benchapp.NewNexGouApp(config, store)
	return app.Handler(), func() { _ = store.Close() }
}

func requestJSON(t *testing.T, handler http.Handler, method string, path string, payload any, expectedStatus int) map[string]any {
	t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, path, body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != expectedStatus {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, expectedStatus, recorder.Body.String())
	}

	response := map[string]any{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response for %s %s: %v; body=%s", method, path, err, recorder.Body.String())
	}
	return response
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
