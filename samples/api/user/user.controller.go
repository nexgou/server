package user

import (
	"time"

	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/middleware"
)

// ── Sample Guards & Interceptors ──────────────────────────────────────────────

// AuthGuard checks for the presence of an Authorization header.
type AuthGuard struct{}

func (g *AuthGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
	return ctx.Header("Authorization") != "", nil
}

// RoleGuard is a placeholder guard for role-based access control.
type RoleGuard struct{}

func (g *RoleGuard) CanActivate(ctx *nexgou.Context) (bool, error) {
	return true, nil
}

// LogInterceptor is a placeholder interceptor for request logging.
type LogInterceptor struct{}

func (i *LogInterceptor) Intercept(ctx *nexgou.Context, next nexgou.HandlerFunc) error {
	return next(ctx)
}

// ── Controller ────────────────────────────────────────────────────────────────

// UserController handles all HTTP requests for the /users resource.
type UserController struct {
	userService *UserService
}

// NewUserController creates a UserController with injected dependencies.
// The IoC container resolves *UserService automatically.
func NewUserController(s *UserService) *UserController {
	return &UserController{userService: s}
}

// Register declares the routes handled by this controller.
func (c *UserController) Register() []nexgou.Route {
	return []nexgou.Route{
		// Standard route — global middleware (RateLimit, Timeout, BodyLimit) applies.
		nexgou.Get("/users", c.FindAll).
			Guard(&AuthGuard{}, &RoleGuard{}).
			Intercept(&LogInterceptor{}).
			Version("v1"),

		// Sensitive route: tighter per-route rate limit (10 req/min per IP),
		// a short timeout, and a small body cap — on top of the global limits.
		nexgou.Post("/users", c.Create).
			Guard(
				&AuthGuard{},
				&middleware.RateLimitGuard{Max: 10, Window: time.Minute},
			).
			Intercept(
				&middleware.TimeoutInterceptor{Duration: 10 * time.Second},
				&middleware.BodyLimitInterceptor{MaxBytes: 64 << 10}, // 64 KB
			).
			Version("v1"),

		// Route with a generous timeout and body limit for heavy payloads.
		nexgou.Get("/users/:id", c.FindOne).
			Intercept(
				&middleware.TimeoutInterceptor{Duration: 5 * time.Second},
			).
			Version("v1"),
	}
}

// FindAll handles GET /users
func (c *UserController) FindAll(ctx *nexgou.Context) error {
	return ctx.JSON(200, c.userService.FindAll())
}

// Create handles POST /users
func (c *UserController) Create(ctx *nexgou.Context) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.Body(&body); err != nil {
		return nexgou.BadRequestException("invalid request body")
	}
	if body.Name == "" {
		return nexgou.BadRequestException("name is required")
	}
	return ctx.JSON(201, c.userService.Create(body.Name))
}

// FindOne handles GET /users/:id
func (c *UserController) FindOne(ctx *nexgou.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(200, c.userService.FindOne(id))
}
