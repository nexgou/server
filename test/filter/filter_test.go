package filter_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/filter"
)

func TestHttpExceptionFilter_HttpException(t *testing.T) {
	f := &filter.HttpExceptionFilter{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := common.NewContext(w, r, nil)

	err := f.Catch(common.NewNotFoundException("user not found"), ctx)
	if err != nil {
		t.Fatalf("Catch: unexpected error: %v", err)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), "user not found") {
		t.Errorf("body: %q does not contain message", w.Body.String())
	}
}

func TestHttpExceptionFilter_GenericError(t *testing.T) {
	f := &filter.HttpExceptionFilter{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := common.NewContext(w, r, nil)

	err := f.Catch(&genericErr{}, ctx)
	if err != nil {
		t.Fatalf("Catch generic: unexpected error: %v", err)
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("generic status: got %d, want 500", w.Code)
	}
}

type genericErr struct{}

func (e *genericErr) Error() string { return "something went wrong" }
