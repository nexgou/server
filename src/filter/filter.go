package filter

import "github.com/nexgou/server/src/common"

// HttpExceptionFilter is the built-in exception filter.
// It handles HttpException errors with a structured JSON response,
// and falls back to a generic 500 for any other error type.
type HttpExceptionFilter struct{}

func (f *HttpExceptionFilter) Catch(err error, ctx *common.Context) error {
	if ex, ok := err.(*common.HttpException); ok {
		return ctx.JSON(ex.Status, common.H{
			"statusCode": ex.Status,
			"message":    ex.Message,
		})
	}
	return ctx.JSON(500, common.H{
		"statusCode": 500,
		"message":    "Internal Server Error",
	})
}
