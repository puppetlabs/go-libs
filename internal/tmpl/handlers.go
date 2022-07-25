package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloWorld is a basic handler setup to illustrate what a handler would look like.
func HelloWorld() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world")
	}
}
