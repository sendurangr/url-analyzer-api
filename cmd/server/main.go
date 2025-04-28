package main

import (
	"github.com/sendurangr/url-analyzer-api/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routes.SetupRouters(r)
	r.Run(":8080")
}
