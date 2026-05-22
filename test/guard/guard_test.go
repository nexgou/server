package guard_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/guard"
)

type allowGuard struct{}

func (g *allowGuard) CanActivate(_ *common.Context) (bool, error) { return true, nil }

type denyGuard struct{}

func (g *denyGuard) CanActivate(_ *common.Context) (bool, error) { return false, nil }

type errGuard struct{}

func (g *errGuard) CanActivate(_ *common.Context) (bool, error) {
	return false, common.NewUnauthorizedException("no token")
}

func newCtx() *common.Context {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	return common.NewContext(httptest.NewRecorder(), r, nil)
}

func TestExecute_AllAllow(t *testing.T) {
	err := guard.Execute(newCtx(), &allowGuard{}, &allowGuard{})
	if err != nil {
		t.Errorf("all allow: unexpected error: %v", err)
	}
}

func TestExecute_OneDenies(t *testing.T) {
	err := guard.Execute(newCtx(), &allowGuard{}, &denyGuard{})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	if ex, ok := err.(*common.HttpException); !ok || ex.Status != 403 {
		t.Errorf("deny: got %v, want 403 Forbidden", err)
	}
}

func TestExecute_GuardError(t *testing.T) {
	err := guard.Execute(newCtx(), &errGuard{})
	if err == nil {
		t.Fatal("expected guard error to propagate")
	}
	if ex, ok := err.(*common.HttpException); !ok || ex.Status != 401 {
		t.Errorf("guard error: got %v, want 401", err)
	}
}

func TestExecute_NoGuards(t *testing.T) {
	err := guard.Execute(newCtx())
	if err != nil {
		t.Errorf("no guards: unexpected error: %v", err)
	}
}
