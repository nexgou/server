package auth

import (
	nexgou "github.com/nexgou/server"
	"github.com/nexgou/server/src/module/validation"
)

// LoginDTO is the request body for POST /v1/auth/login.
type LoginDTO struct {
	Username string `json:"username" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthController handles HTTP requests for /auth routes.
type AuthController struct {
	authService *AuthService
	validation  *validation.ValidationService
}

// NewAuthController creates an AuthController with injected dependencies.
func NewAuthController(svc *AuthService, v *validation.ValidationService) *AuthController {
	return &AuthController{authService: svc, validation: v}
}

// Register declares the routes handled by this controller.
func (c *AuthController) Register() []nexgou.Route {
	return []nexgou.Route{
		nexgou.Post("/auth/login", c.Login).Version("v1"),
	}
}

// Login handles POST /v1/auth/login
func (c *AuthController) Login(ctx *nexgou.Context) error {
	var dto LoginDTO
	if err := ctx.Body(&dto); err != nil {
		return nexgou.BadRequestException("invalid request body")
	}
	if err := c.validation.ValidateStruct(dto); err != nil {
		return nexgou.BadRequestException(err.Error())
	}
	token, err := c.authService.Login(dto.Username, dto.Password)
	if err != nil {
		return nexgou.InternalServerErrorException("could not issue token")
	}
	if token == "" {
		return nexgou.UnauthorizedException("invalid credentials")
	}
	return ctx.JSON(200, map[string]string{"access_token": token})
}
