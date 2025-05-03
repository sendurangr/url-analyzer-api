package utils

import "github.com/gin-gonic/gin"

func RespondWithError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, gin.H{
		"message": message,
	})
}
