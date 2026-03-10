package middleware

import (
	"restaurant/internal/adapter/handler/translator"

	"github.com/gin-gonic/gin"
)

// Error returns a middleware handles all stored errors in [gin.Context].
func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) == 0 {
			return
		}

		resp := translator.TranslateError(ctx.Errors.Last())
		ctx.AbortWithStatusJSON(resp.Status, resp)
	}
}
