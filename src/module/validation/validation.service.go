// Package validation provides struct and value validation for Nexgou applications.
//
// It wraps github.com/go-playground/validator/v10 and exposes both a ValidationService
// for programmatic use and a ValidationPipe for automatic request body validation.
//
// Usage — validate a request body in a handler:
//
//	func (c *UserController) Create(ctx *nexgou.Context) error {
//	    var dto CreateUserDTO
//	    if err := ctx.Body(&dto); err != nil {
//	        return nexgou.BadRequestException("invalid JSON")
//	    }
//	    if err := c.validation.ValidateStruct(dto); err != nil {
//	        return nexgou.BadRequestException(err.Error())
//	    }
//	    ...
//	}
package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nexgou/server/src/logger"
)

// ValidationError describes a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a slice of field errors returned by ValidateStruct.
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// ValidationService validates Go structs using field tags.
//
// Supported tags (from go-playground/validator):
//
//	required, min, max, len, email, url, uuid, oneof, gte, lte, gt, lt, ...
//
// Example struct:
//
//	type CreateUserDTO struct {
//	    Name  string `validate:"required,min=2,max=100"`
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"gte=18,lte=120"`
//	}
type ValidationService struct {
	v   *validator.Validate
	log *logger.ScopedLogger
}

// NewValidationService creates a new ValidationService.
// Depends on *logger.LoggerService.
func NewValidationService(log *logger.LoggerService) *ValidationService {
	v := validator.New()
	svc := &ValidationService{
		v:   v,
		log: log.WithContext("ValidationService"),
	}
	svc.log.Info("initialized")
	return svc
}

// ValidateStruct validates a struct using its `validate` tags.
// Returns a ValidationErrors slice if any fields fail validation, nil otherwise.
func (s *ValidationService) ValidateStruct(obj any) error {
	err := s.v.Struct(obj)
	if err == nil {
		return nil
	}
	var verrs ValidationErrors
	for _, fe := range err.(validator.ValidationErrors) {
		verrs = append(verrs, &ValidationError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: humanizeError(fe),
		})
	}
	return verrs
}

// ValidateVar validates a single value against the given tag expression.
//
//	err := svc.ValidateVar("user@example.com", "required,email")
func (s *ValidationService) ValidateVar(value any, tag string) error {
	return s.v.Var(value, tag)
}

// RegisterValidation registers a custom validation function under the given tag name.
//
//	svc.RegisterValidation("strongpass", func(fl validator.FieldLevel) bool {
//	    return len(fl.Field().String()) >= 12
//	})
func (s *ValidationService) RegisterValidation(tag string, fn validator.Func) error {
	return s.v.RegisterValidation(tag, fn)
}

// humanizeError converts a validator.FieldError into a human-readable message.
func humanizeError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
