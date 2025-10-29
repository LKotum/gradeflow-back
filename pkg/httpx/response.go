package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gradeflow/internal/domain/dto/response"
)

// WriteError writes a unified error response.
func WriteError(ctx *gin.Context, status int, code, message string, details map[string]any) {
	ctx.AbortWithStatusJSON(status, response.ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// WriteData writes successful JSON response.
func WriteData(ctx *gin.Context, status int, payload any) {
	ctx.JSON(status, payload)
}

// WriteNoContent writes HTTP 204 without body.
func WriteNoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}
