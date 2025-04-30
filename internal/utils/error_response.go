package utils

import "github.com/gin-gonic/gin"

func RespondWithError(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, gin.H{"message": msg})
}
