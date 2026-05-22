package common

import "fmt"

// HttpException represents a structured HTTP error with a status code and message.
type HttpException struct {
	Status  int
	Message string
}

func (e *HttpException) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

// NewHttpException creates a generic HTTP exception.
func NewHttpException(status int, message string) *HttpException {
	return &HttpException{Status: status, Message: message}
}

// NewBadRequestException creates a 400 Bad Request exception.
func NewBadRequestException(message string) *HttpException {
	return &HttpException{Status: 400, Message: message}
}

// NewUnauthorizedException creates a 401 Unauthorized exception.
func NewUnauthorizedException(message string) *HttpException {
	return &HttpException{Status: 401, Message: message}
}

// NewForbiddenException creates a 403 Forbidden exception.
func NewForbiddenException(message string) *HttpException {
	return &HttpException{Status: 403, Message: message}
}

// NewNotFoundException creates a 404 Not Found exception.
func NewNotFoundException(message string) *HttpException {
	return &HttpException{Status: 404, Message: message}
}

// NewInternalServerErrorException creates a 500 Internal Server Error exception.
func NewInternalServerErrorException(message string) *HttpException {
	return &HttpException{Status: 500, Message: message}
}
