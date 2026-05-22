// Package nexgoutest provides lightweight test helpers for Nexgou applications.
//
// It exposes two layers of testing support:
//
//  1. Unit testing via NewContext — create an isolated *common.Context with
//     mocked request data, without spinning up an HTTP server.
//
//  2. Integration testing via NewSuite — launch a real httptest.Server backed
//     by a Nexgou app and use the fluent RequestBuilder / ResponseAssertion API.
//
// Example unit test:
//
//	func TestFindAll(t *testing.T) {
//	    ctx := nexgoutest.NewContext(t, nexgoutest.WithMethod("GET"), nexgoutest.WithPath("/users"))
//	    svc := &UserService{}
//	    ctrl := NewUserController(svc)
//	    if err := ctrl.FindAll(ctx); err != nil {
//	        t.Fatalf("unexpected error: %v", err)
//	    }
//	    ctx.Assert(t).Status(200).BodyContains("Alice")
//	}
//
// Example integration test:
//
//	func TestIntegration(t *testing.T) {
//	    suite := nexgoutest.NewSuite(t, AppModule)
//	    defer suite.Close()
//	    suite.GET("/v1/users").
//	        Header("Authorization", "Bearer token").
//	        Do(t).
//	        Status(200).
//	        BodyContains("Alice")
//	}
package nexgoutest
