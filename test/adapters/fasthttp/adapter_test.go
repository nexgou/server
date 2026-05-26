package fasthttpadapter_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	fasthttpadapter "github.com/nexgou/server/src/adapters/fasthttp"
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestFastHTTPAdapterServesNexGouHandler(t *testing.T) {
	router := router.New()
	router.Add(common.Route{Method: http.MethodGet, Path: "/health", Handler: func(ctx *common.Context) error {
		return ctx.JSON(http.StatusOK, common.H{"status": "ok"})
	}})

	listener := fasthttputil.NewInmemoryListener()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- fasthttp.Serve(listener, fasthttpadapter.NewHandler(router))
	}()

	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network string, address string) (net.Conn, error) {
		return listener.Dial()
	}}}

	response, err := client.Get("http://nexgou/health")
	if err != nil {
		t.Fatalf("GET /health through fasthttp adapter: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if string(body) != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q, want health JSON", string(body))
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	if err := <-serverErrors; err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		t.Fatalf("fasthttp server returned unexpected error: %v", err)
	}
}
