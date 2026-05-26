package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/router"
)

func benchmarkHandler(ctx *common.Context) error {
	return nil
}

func runRouterBenchmark(b *testing.B, router *router.Router, method string, path string) {
	b.Helper()
	b.ReportAllocs()

	request := httptest.NewRequest(method, path, nil)
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		router.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func BenchmarkRouterStaticRoute(b *testing.B) {
	router := router.New()
	router.Add(common.Route{Method: http.MethodGet, Path: "/tasks", Handler: benchmarkHandler})
	router.Add(common.Route{Method: http.MethodGet, Path: "/health", Handler: benchmarkHandler})

	runRouterBenchmark(b, router, http.MethodGet, "/tasks")
}

func BenchmarkRouterParamRoute(b *testing.B) {
	router := router.New()
	router.Add(common.Route{Method: http.MethodGet, Path: "/tasks/:id", Handler: benchmarkHandler})

	runRouterBenchmark(b, router, http.MethodGet, "/tasks/123")
}

func BenchmarkRouterNotFound(b *testing.B) {
	router := router.New()
	for _, path := range []string{"/tasks", "/users", "/metrics", "/events"} {
		router.Add(common.Route{Method: http.MethodGet, Path: path, Handler: benchmarkHandler})
	}

	runRouterBenchmark(b, router, http.MethodGet, "/missing")
}

func BenchmarkRouterHighCardinalityStatic(b *testing.B) {
	for _, count := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("routes_%d", count), func(b *testing.B) {
			router := router.New()
			for index := 0; index < count; index++ {
				path := fmt.Sprintf("/resources/%d", index)
				router.Add(common.Route{Method: http.MethodGet, Path: path, Handler: benchmarkHandler})
			}

			target := fmt.Sprintf("/resources/%d", count-1)
			runRouterBenchmark(b, router, http.MethodGet, target)
		})
	}
}

func BenchmarkRouterFullPipeline(b *testing.B) {
	router := router.New()
	router.Use(func(next common.HandlerFunc) common.HandlerFunc {
		return func(ctx *common.Context) error {
			return next(ctx)
		}
	})
	router.Add(common.Route{Method: http.MethodGet, Path: "/tasks/:id", Handler: benchmarkHandler}.
		Guard(benchmarkGuard{}).
		Pipe(benchmarkPipe{}).
		Intercept(benchmarkInterceptor{}))

	runRouterBenchmark(b, router, http.MethodGet, "/tasks/123")
}

type benchmarkGuard struct{}

func (guard benchmarkGuard) CanActivate(ctx *common.Context) (bool, error) {
	return true, nil
}

type benchmarkPipe struct{}

func (pipe benchmarkPipe) Transform(value string) (any, error) {
	return value, nil
}

type benchmarkInterceptor struct{}

func (interceptor benchmarkInterceptor) Intercept(ctx *common.Context, next common.HandlerFunc) error {
	return next(ctx)
}
