package task

import (
	nexgou "github.com/nexgou/server"
	jwtmod "github.com/nexgou/server/src/module/jwt"
	"github.com/nexgou/server/src/module/validation"
)

// CreateTaskDTO is the request body for POST /v1/tasks.
type CreateTaskDTO struct {
	Title string `json:"title" validate:"required,min=1,max=200"`
}

// TaskController handles HTTP requests for the /tasks resource.
type TaskController struct {
	service    *TaskService
	jwt        *jwtmod.JwtService
	validation *validation.ValidationService
}

// NewTaskController creates a TaskController with injected dependencies.
func NewTaskController(svc *TaskService, jwt *jwtmod.JwtService, v *validation.ValidationService) *TaskController {
	return &TaskController{service: svc, jwt: jwt, validation: v}
}

// Register declares the routes handled by this controller.
func (c *TaskController) Register() []nexgou.Route {
	guard := &jwtmod.JwtGuard{Jwt: c.jwt}
	return []nexgou.Route{
		nexgou.Get("/tasks", c.FindAll).Guard(guard).Version("v1"),
		nexgou.Post("/tasks", c.Create).Guard(guard).Version("v1"),
		nexgou.Patch("/tasks/:id/complete", c.Complete).Guard(guard).Version("v1"),
		nexgou.Delete("/tasks/:id", c.Delete).Guard(guard).Version("v1"),
	}
}

// currentUser extracts the username from the verified JWT.
func (c *TaskController) currentUser(ctx *nexgou.Context) string {
	tokenStr := jwtmod.ExtractFromHeader(ctx.Header("Authorization"))
	claims, err := c.jwt.Verify(tokenStr)
	if err != nil || claims.Data == nil {
		return "anonymous"
	}
	if u, ok := claims.Data["username"].(string); ok {
		return u
	}
	return "anonymous"
}

// FindAll handles GET /v1/tasks
func (c *TaskController) FindAll(ctx *nexgou.Context) error {
	userID := c.currentUser(ctx)
	tasks, err := c.service.FindAll(userID)
	if err != nil {
		return nexgou.InternalServerErrorException("failed to fetch tasks")
	}
	return ctx.JSON(200, tasks)
}

// Create handles POST /v1/tasks
func (c *TaskController) Create(ctx *nexgou.Context) error {
	var dto CreateTaskDTO
	if err := ctx.Body(&dto); err != nil {
		return nexgou.BadRequestException("invalid request body")
	}
	if err := c.validation.ValidateStruct(dto); err != nil {
		return nexgou.BadRequestException(err.Error())
	}
	userID := c.currentUser(ctx)
	t, err := c.service.Create(dto.Title, userID)
	if err != nil {
		return nexgou.InternalServerErrorException("failed to create task")
	}
	return ctx.JSON(201, t)
}

// Complete handles PATCH /v1/tasks/:id/complete
func (c *TaskController) Complete(ctx *nexgou.Context) error {
	id := ctx.Param("id")
	t, err := c.service.Complete(id)
	if err != nil {
		return nexgou.NotFoundException("task not found")
	}
	return ctx.JSON(200, t)
}

// Delete handles DELETE /v1/tasks/:id
func (c *TaskController) Delete(ctx *nexgou.Context) error {
	id := ctx.Param("id")
	if err := c.service.Delete(id); err != nil {
		return nexgou.NotFoundException("task not found")
	}
	return ctx.JSON(200, map[string]string{"status": "deleted"})
}
